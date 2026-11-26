// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package shortcut

import (
	"context"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	shortcutmodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/shortcut"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ShortcutCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShortcutCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShortcutCreateLogic {
	return &ShortcutCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ShortcutCreate 为当前用户绑定菜单资源为快捷入口。
func (l *ShortcutCreateLogic) ShortcutCreate(req *types.ShortcutCreateReq) (resp *types.IdResp, err error) {
	// 1. 登录与入参
	uid := authctx.UserId(l.ctx)
	if uid == "" {
		return nil, errorx.New(401, "未登录或登录已失效")
	}
	resourceId := strings.TrimSpace(req.ResourceId)
	if resourceId == "" {
		return nil, errorx.New(400, "请选择菜单资源")
	}
	sort, err := convert.ParseSort(req.Sort, 0)
	if err != nil {
		return nil, err
	}

	// 2. 资源必须是启用中的菜单（禁止目录/按钮）
	res, err := l.svcCtx.Repo.Resource.FindById(l.ctx, resourceId)
	if err != nil {
		return nil, errorx.New(400, "资源不存在")
	}
	if res.Type != "menu" {
		return nil, errorx.New(400, "只能选择菜单类型资源")
	}
	if res.Status != 1 {
		return nil, errorx.New(400, "资源已停用")
	}

	// 3. 上限与去重
	n, err := l.svcCtx.Repo.Shortcut.CountByUser(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	if n >= shortcutmodel.MaxPerUser {
		return nil, errorx.New(400, "快捷入口已达上限")
	}
	if _, err := l.svcCtx.Repo.Shortcut.FindByUserAndResource(l.ctx, uid, resourceId); err == nil {
		return nil, errorx.New(400, "该菜单已添加")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 4. 只落 resource_id
	id, err := l.svcCtx.IdGen.Next()
	if err != nil {
		return nil, err
	}
	if err := l.svcCtx.Repo.Shortcut.Create(l.ctx, &shortcutmodel.SysUserShortcut{
		Id: id, UserId: uid, ResourceId: resourceId, Sort: sort, CreatedBy: uid, UpdatedBy: uid,
	}); err != nil {
		return nil, err
	}
	operlog.SetSummary(l.ctx, "添加快捷入口「"+res.Name+"」")
	return &types.IdResp{Id: id}, nil
}
