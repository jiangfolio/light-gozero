package user

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/logic/datascope"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type UserPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserPageLogic {
	return &UserPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *UserPageLogic) UserPage(req *types.UserPageReq) (resp *types.UserPageResp, err error) {
	uid := authctx.UserId(l.ctx)
	if err := l.svcCtx.RbacGuard.Require(l.ctx, uid, perm.SystemUserPage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
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

	// 1. 解析数据范围并与请求 org_id 求交
	scope, err := l.svcCtx.RbacGuard.ResolveDataScope(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	scopeOrgIds, selfUserId, empty := datascope.ForUserPage(scope, uid, req.OrgId)
	if empty {
		return &types.UserPageResp{
			Items: []types.UserItem{}, Page: convert.Int64ToString(int64(page)),
			PageSize: convert.Int64ToString(int64(pageSize)), Total: "0",
		}, nil
	}
	filterOrgId := req.OrgId
	if scope != nil && !scope.All && !scope.SelfOnly {
		// 已在 scopeOrgIds 中求交，避免重复 org_id= 与 IN 冲突放大
		filterOrgId = ""
	}

	list, total, err := l.svcCtx.Repo.User.Page(l.ctx, req.Username, filterOrgId, status, hasStatus, scopeOrgIds, selfUserId, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.UserItem, 0, len(list))
	for i := range list {
		items = append(items, *toUserItem(&list[i]))
	}
	return &types.UserPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)), Total: convert.Int64ToString(total),
	}, nil
}
