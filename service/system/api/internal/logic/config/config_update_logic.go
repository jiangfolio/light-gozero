package config

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ConfigUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfigUpdateLogic {
	return &ConfigUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ConfigUpdate 仅允许改 value/remark/status，禁止改 key/value_type。
func (l *ConfigUpdateLogic) ConfigUpdate(req *types.ConfigUpdateReq) error {
	// 1. 校验操作权限与入参
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemConfigUpdate); err != nil {
		return err
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Config.FindById(l.ctx, req.Id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorx.New(404, "配置不存在")
		}
		return err
	}

	// 2. 按 value_type 与业务键边界校验新值
	value := strings.TrimSpace(req.ConfigValue)
	if err := validateConfigValue(row.ValueType, value); err != nil {
		return err
	}
	if err := validateConfigBounds(row.ConfigKey, row.ValueType, value); err != nil {
		return err
	}
	row.ConfigValue = value
	if req.Remark != "" {
		row.Remark = req.Remark
	}
	if req.Status != "" {
		status, err := convert.ParseStatus(req.Status, row.Status)
		if err != nil {
			return err
		}
		row.Status = status
	}
	row.UpdatedBy = authctx.UserId(l.ctx)

	// 3. 落库后失效缓存，保证热生效
	if err := l.svcCtx.Repo.Config.Save(l.ctx, row); err != nil {
		return err
	}
	_ = l.svcCtx.AppConfig.Invalidate(l.ctx)
	operlog.SetSummary(l.ctx, "更新配置「"+row.ConfigKey+"」")
	return nil
}

func validateConfigValue(valueType, value string) error {
	switch valueType {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return errorx.New(400, "配置值须为整数")
		}
	case "bool":
		if value != "true" && value != "false" && value != "1" && value != "0" {
			return errorx.New(400, "配置值须为布尔值")
		}
	case "string", "json", "":
		// 允许空串以外的任意文本；json 仅存文本，不强制解析
	default:
		return errorx.New(400, "不支持的配置值类型")
	}
	return nil
}

func validateConfigBounds(key, valueType, value string) error {
	if valueType != "int" {
		return nil
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return errorx.New(400, "配置值须为整数")
	}
	switch key {
	case appcfg.KeyTokenTTLSeconds:
		if n < 60 || n > 604800 {
			return errorx.New(400, "会话有效期须在 60～604800 秒")
		}
	case appcfg.KeyMaxFailAttempts:
		if n < 1 || n > 20 {
			return errorx.New(400, "失败锁定阈值须在 1～20")
		}
	case appcfg.KeyLockMinutes:
		if n < 1 || n > 1440 {
			return errorx.New(400, "锁定时长须在 1～1440 分钟")
		}
	case appcfg.KeyPasswordMinLen:
		if n < 6 || n > 64 {
			return errorx.New(400, "密码最小长度须在 6～64")
		}
	case appcfg.KeyPasswordExpireDays:
		if n < 0 || n > 3650 {
			return errorx.New(400, "密码过期天数须在 0～3650")
		}
	case appcfg.KeyPasswordHistoryCount:
		if n < 0 || n > 24 {
			return errorx.New(400, "历史密码条数须在 0～24")
		}
	case appcfg.KeyCaptchaTTLSeconds:
		if n < 30 || n > 600 {
			return errorx.New(400, "验证码有效期须在 30～600 秒")
		}
	case appcfg.KeyCaptchaLength:
		if n < 4 || n > 8 {
			return errorx.New(400, "验证码位数须在 4～8")
		}
	}
	return nil
}
