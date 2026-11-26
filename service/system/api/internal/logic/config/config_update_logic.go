package config

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/batchutil"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model"
	configmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/config"
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

// ConfigUpdate 按 config_key 批量更新 config_value，禁止改 key/value_type。
func (l *ConfigUpdateLogic) ConfigUpdate(req *types.ConfigUpdateReq) error {
	// 1. 校验权限与 items：非空、条数上限、key 去重
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemConfigUpdate); err != nil {
		return err
	}
	if len(req.Items) == 0 {
		return errorx.New(400, "缺少 items")
	}
	if len(req.Items) > batchutil.MaxRequestIDs {
		return errorx.New(400, "单次最多处理 100 条")
	}
	seen := make(map[string]struct{}, len(req.Items))
	keys := make([]string, 0, len(req.Items))
	valueByKey := make(map[string]string, len(req.Items))
	for _, it := range req.Items {
		key := strings.TrimSpace(it.ConfigKey)
		if key == "" {
			return errorx.New(400, "config_key 不能为空")
		}
		if _, ok := seen[key]; ok {
			return errorx.New(400, "配置键重复："+key)
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
		valueByKey[key] = strings.TrimSpace(it.ConfigValue)
	}

	// 2. 按 key 批量加载；缺 key 拒绝（禁止隐式创建）
	rows, err := l.svcCtx.Repo.Config.ListByKeys(l.ctx, keys)
	if err != nil {
		return err
	}
	rowByKey := make(map[string]*configmodel.SysConfig, len(rows))
	for i := range rows {
		rowByKey[rows[i].ConfigKey] = &rows[i]
	}
	for _, key := range keys {
		row, ok := rowByKey[key]
		if !ok {
			return errorx.New(404, "配置不存在："+key)
		}
		// 安全分组须走 /config/security/update
		if isSecurityConfigGroup(row.GroupName) {
			return errorx.New(403, "安全配置请使用安全配置接口更新："+key)
		}
	}

	// 3. 按 value_type 与业务边界校验，组装待写行
	op := authctx.UserId(l.ctx)
	toSave := make([]*configmodel.SysConfig, 0, len(keys))
	for _, key := range keys {
		row := rowByKey[key]
		value := valueByKey[key]
		if err := validateConfigValue(row.ValueType, value); err != nil {
			return err
		}
		if err := validateConfigBounds(row.ConfigKey, row.ValueType, value); err != nil {
			return err
		}
		row.ConfigValue = value
		row.UpdatedBy = op
		toSave = append(toSave, row)
	}

	// 4. 短事务批量落库；事务外失效缓存
	if err := l.svcCtx.Repo.Transaction(l.ctx, func(tx *model.Repo) error {
		for _, row := range toSave {
			if err := tx.Config.Save(l.ctx, row); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	_ = l.svcCtx.AppConfig.Invalidate(l.ctx)
	operlog.SetSummary(l.ctx, "批量更新配置 "+convert.Int64ToString(int64(len(toSave)))+" 项")
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
