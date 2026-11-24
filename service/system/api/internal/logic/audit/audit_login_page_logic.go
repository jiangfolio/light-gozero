package audit

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/jiangbyte/light-gozero/common/auditlabel"
	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/pageutil"
	"github.com/jiangbyte/light-gozero/common/perm"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

type AuditLoginPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditLoginPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditLoginPageLogic {
	return &AuditLoginPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AuditLoginPageLogic) AuditLoginPage(req *types.AuditLoginPageReq) (resp *types.AuditLoginPageResp, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemAuditLoginPage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list, total, err := l.svcCtx.Repo.Audit.PageLoginLogs(l.ctx, req.Username, req.Result, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
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
