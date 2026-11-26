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
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type ShortcutUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShortcutUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShortcutUpdateLogic {
	return &ShortcutUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ShortcutUpdate 更新本人快捷入口的资源或排序。
func (l *ShortcutUpdateLogic) ShortcutUpdate(req *types.ShortcutUpdateReq) error {
	// 1. 登录与归属
	uid := authctx.UserId(l.ctx)
	if uid == "" {
		return errorx.New(401, "未登录或登录已失效")
	}
	if req.Id == "" {
		return errorx.New(400, "缺少 id")
	}
	row, err := l.svcCtx.Repo.Shortcut.FindById(l.ctx, req.Id)
	if err != nil {
		return errorx.New(404, "快捷入口不存在")
	}
	if row.UserId != uid {
		return errorx.New(403, "无权修改他人快捷入口")
	}

	// 2. 可选更换资源
	title := ""
	if req.ResourceId != "" {
		resourceId := strings.TrimSpace(req.ResourceId)
		res, err := l.svcCtx.Repo.Resource.FindById(l.ctx, resourceId)
		if err != nil {
			return errorx.New(400, "资源不存在")
		}
		if res.Type != "menu" {
			return errorx.New(400, "只能选择菜单类型资源")
		}
		if res.Status != 1 {
			return errorx.New(400, "资源已停用")
		}
		if resourceId != row.ResourceId {
			if _, err := l.svcCtx.Repo.Shortcut.FindByUserAndResource(l.ctx, uid, resourceId); err == nil {
				return errorx.New(400, "该菜单已添加")
			} else if err != gorm.ErrRecordNotFound {
				return err
			}
		}
		row.ResourceId = resourceId
		title = res.Name
	}
	if req.Sort != "" {
		v, err := convert.ParseSort(req.Sort, row.Sort)
		if err != nil {
			return err
		}
		row.Sort = v
	}
	row.UpdatedBy = uid

	// 3. 落库
	if err := l.svcCtx.Repo.Shortcut.Update(l.ctx, row); err != nil {
		return err
	}
	if title == "" {
		if res, err := l.svcCtx.Repo.Resource.FindById(l.ctx, row.ResourceId); err == nil {
			title = res.Name
		}
	}
	operlog.SetSummary(l.ctx, "更新快捷入口「"+title+"」")
	return nil
}
