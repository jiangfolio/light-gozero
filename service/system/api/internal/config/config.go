// Code scaffolded by goctl. Safe to edit.
package config

import (
	"github.com/jiangbyte/light-gozero/common/storage"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	IdGen struct {
		Mode string
	}
	PwdCipher struct {
		PrivateKey string `json:",env=SYSTEM_PWD_PRIVATE_KEY"`
	}
	Mysql struct {
		DataSource string `json:",env=SYSTEM_MYSQL_DATASOURCE"`
	}
	Redis redis.RedisConf
	// TrustProxy 为 true 时才信任 X-Forwarded-For / X-Real-IP。
	TrustProxy bool `json:",optional"`
	// RateLimit 公开接口周期限流（秒内配额）。
	RateLimit struct {
		PeriodSeconds  int `json:",default=60"`
		LoginQuota     int `json:",default=20"`
		CaptchaQuota   int `json:",default=60"`
		PublicKeyQuota int `json:",default=60"`
	} `json:",optional"`
	// Storage 对象存储（local 分片 / s3 直传），见 common/storage。
	Storage storage.Config `json:",optional"`
}
