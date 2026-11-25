package dict

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type DictItemPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDictItemPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictItemPageLogic {
	return &DictItemPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *DictItemPageLogic) DictItemPage(req *types.DictItemPageReq) (resp *types.DictItemPageResp, err error) {
	// 1. 校验权限；右侧列表必须绑定左侧选中的类型
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemDictPage); err != nil {
		return nil, err
	}
	if req.ParentId == "" {
		return nil, errorx.New(400, "缺少 parent_id")
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	parent, err := l.svcCtx.Repo.Dict.FindById(l.ctx, req.ParentId)
	if err != nil {
		return nil, errorx.New(400, "上级字典不存在")
	}
	if parent.ParentId != "" {
		return nil, errorx.New(400, "parent_id 必须是字典类型")
	}
	var status int64
	hasStatus := false
	if req.Status != "" {
		status, err = convert.Int64FromString(req.Status)
		if err != nil {
			return nil, err
		}
		hasStatus = true
	}
	// 2. 按类型 id 分页查项
	list, total, err := l.svcCtx.Repo.Dict.PageItems(l.ctx, req.ParentId, req.Label, req.Value, status, hasStatus, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.DictInfo, 0, len(list))
	for i := range list {
		items = append(items, *toDictInfo(&list[i]))
	}
	return &types.DictItemPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)),
		PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}
