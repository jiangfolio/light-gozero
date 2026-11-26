package config

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

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

type SecurityConfigUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSecurityConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SecurityConfigUpdateLogic {
	return &SecurityConfigUpdateLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// SecurityConfigUpdate 仅允许更新 auth/captcha 分组配置。
func (l *SecurityConfigUpdateLogic) SecurityConfigUpdate(req *types.ConfigUpdateReq) error {
	// 1. 校验权限与 items
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemSecurityConfigUpdate); err != nil {
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

	// 2. 批量加载并校验均属安全分组
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
		if !isSecurityConfigGroup(row.GroupName) {
			return errorx.New(403, "非安全配置不可由此接口更新："+key)
		}
	}

	// 3. 校验值并组装待写行
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

	// 4. 短事务落库；事务外刷缓存
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
	operlog.SetSummary(l.ctx, "更新安全配置 "+convert.Int64ToString(int64(len(toSave)))+" 项")
	return nil
}
