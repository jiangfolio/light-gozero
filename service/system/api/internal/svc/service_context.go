package svc

import (
	"context"
	"fmt"
	"time"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/captcha"
	"github.com/jiangbyte/light-gozero/common/idgen"
	"github.com/jiangbyte/light-gozero/common/pwdcipher"
	"github.com/jiangbyte/light-gozero/common/rbac"
	"github.com/jiangbyte/light-gozero/common/session"
	"github.com/jiangbyte/light-gozero/common/storage"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/config"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/middleware"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/audit"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ServiceContext struct {
	Config       config.Config
	Auth         rest.Middleware
	OperLog      rest.Middleware
	RateLimit    rest.Middleware
	IdGen        idgen.Generator
	SessionStore *session.Store
	CaptchaStore *captcha.Store
	AppConfig    *appcfg.Service
	PwdCipher    *pwdcipher.Cipher
	RbacGuard    *rbac.Guard
	Repo         *model.Repo
	FileStore    storage.Store
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 1. ID 策略
	gen, err := idgen.New(c.IdGen.Mode)
	if err != nil {
		panic(err)
	}
	fmt.Printf("IdGen.Mode=%s\n", gen.Kind())

	// 2. 密码传输密钥
	pwdc, err := pwdcipher.NewFromPrivateKeyHex(c.PwdCipher.PrivateKey)
	if err != nil {
		panic(err)
	}

	// 3. GORM + Redis
	db, err := gorm.Open(mysql.Open(c.Mysql.DataSource), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		panic(err)
	}
	rdb := redis.MustNewRedis(c.Redis)
	repo := model.NewRepo(db)

	// 4. 动态配置
	appCfg := appcfg.New(rdb, repo.Config, 60*time.Second)
	store := session.NewStore(rdb, func(ctx context.Context) time.Duration {
		sec := appCfg.GetInt(ctx, appcfg.KeyTokenTTLSeconds, appcfg.DefaultTokenTTLSeconds)
		return time.Duration(sec) * time.Second
	})
	capStore := captcha.NewStore(rdb,
		func(ctx context.Context) time.Duration {
			sec := appCfg.GetInt(ctx, appcfg.KeyCaptchaTTLSeconds, appcfg.DefaultCaptchaTTLSeconds)
			return time.Duration(sec) * time.Second
		},
		func(ctx context.Context) int {
			return appCfg.GetInt(ctx, appcfg.KeyCaptchaLength, appcfg.DefaultCaptchaLength)
		},
	)
	guard := rbac.NewGuard(rdb, func(ctx context.Context, userId string) ([]string, error) {
		return repo.Resource.ListPermCodesByUser(ctx, userId)
	})
	guard.SetScopeLoader(newScopeLoader(repo))

	authMw := middleware.NewAuthMiddleware(store, func(userId string) (int64, error) {
		u, err := repo.User.FindById(context.Background(), userId)
		if err != nil {
			return 0, err
		}
		return u.Status, nil
	})

	operMw := middleware.NewOperLogMiddleware(func(row *audit.OperLog) {
		if row == nil {
			return
		}
		if row.Id == "" {
			id, err := gen.Next()
			if err != nil {
				return
			}
			row.Id = id
		}
		if err := repo.Audit.CreateOperLog(context.Background(), row); err != nil {
			logx.Errorf("write oper log: %v", err)
		}
	}, gen.Next)

	period := c.RateLimit.PeriodSeconds
	if period <= 0 {
		period = 60
	}
	loginQ := c.RateLimit.LoginQuota
	if loginQ <= 0 {
		loginQ = 20
	}
	captchaQ := c.RateLimit.CaptchaQuota
	if captchaQ <= 0 {
		captchaQ = 60
	}
	pkQ := c.RateLimit.PublicKeyQuota
	if pkQ <= 0 {
		pkQ = 60
	}
	rateMw := middleware.NewRateLimitMiddleware(
		limit.NewPeriodLimit(period, loginQ, rdb, "rl:login:"),
		limit.NewPeriodLimit(period, captchaQ, rdb, "rl:captcha:"),
		limit.NewPeriodLimit(period, pkQ, rdb, "rl:pubkey:"),
	)

	// 5. 对象存储（local / s3），供后续 file 模块注入
	storeCfg := c.Storage
	if storeCfg.Provider == "" {
		storeCfg.Provider = storage.ProviderLocal
	}
	if storeCfg.Provider == storage.ProviderLocal && storeCfg.Local.Root == "" {
		storeCfg.Local.Root = "./data/storage"
	}
	fileStore, err := storage.New(storeCfg)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Storage.Provider=%s AccessMode=%s\n", fileStore.Provider(), fileStore.AccessMode())

	return &ServiceContext{
		Config:       c,
		Auth:         authMw.Handle,
		OperLog:      operMw.Handle,
		RateLimit:    rateMw.Handle,
		IdGen:        gen,
		SessionStore: store,
		CaptchaStore: capStore,
		AppConfig:    appCfg,
		PwdCipher:    pwdc,
		RbacGuard:    guard,
		Repo:         repo,
		FileStore:    fileStore,
	}
}
