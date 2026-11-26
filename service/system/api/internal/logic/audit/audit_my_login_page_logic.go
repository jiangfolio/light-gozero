// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package audit

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/auditlabel"
	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type AuditMyLoginPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditMyLoginPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditMyLoginPageLogic {
	return &AuditMyLoginPageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// AuditMyLoginPage 我的登录日志：登录即可，按会话用户 ID 精确过滤，用户名可改、ID 不可变。
func (l *AuditMyLoginPageLogic) AuditMyLoginPage(req *types.AuditLoginMyPageReq) (resp *types.AuditLoginPageResp, err error) {
	// 1. 仅要求登录态；用户 ID 来自会话，不接收前端传参
	uid := authctx.UserId(l.ctx)
	if uid == "" {
		return nil, errorx.New(401, "未登录或登录已失效")
	}
	// 2. 分页参数与查询（按 user_id 精确匹配）
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list, total, err := l.svcCtx.Repo.Audit.PageLoginLogsByUserId(l.ctx, uid, req.Result, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	// 3. 组装返回
	items := make([]types.AuditLoginItem, 0, len(list))
	for _, row := range list {
		items = append(items, types.AuditLoginItem{
			Id: row.Id, UserId: row.UserId, Username: row.Username, Ip: row.Ip, UserAgent: row.UserAgent,
			Result: row.Result, ResultName: auditlabel.ResultName(row.Result), Message: row.Message,
			CreatedAt: convert.TimeToString(row.CreatedAt),
		})
	}
	return &types.AuditLoginPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)),
		Total: convert.Int64ToString(total),
	}, nil
}
