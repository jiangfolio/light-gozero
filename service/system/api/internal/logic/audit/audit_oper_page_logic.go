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

type AuditOperPageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuditOperPageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditOperPageLogic {
	return &AuditOperPageLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *AuditOperPageLogic) AuditOperPage(req *types.AuditOperPageReq) (resp *types.AuditOperPageResp, err error) {
	if err := l.svcCtx.RbacGuard.Require(l.ctx, authctx.UserId(l.ctx), perm.SystemAuditOperPage); err != nil {
		return nil, err
	}
	page, pageSize, err := pageutil.Parse(req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	list, total, err := l.svcCtx.Repo.Audit.PageOperLogs(l.ctx, req.Username, req.Module, req.Result, pageutil.Offset(page, pageSize), pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]types.AuditOperItem, 0, len(list))
	for _, row := range list {
		items = append(items, types.AuditOperItem{
			Id: row.Id, UserId: row.UserId, Username: row.Username,
			Module: row.Module, Action: row.Action,
			Method: row.Method, Path: row.Path, ReqBody: row.ReqBody, Ip: row.Ip,
			LatencyMs: convert.Int64ToString(row.LatencyMs),
			Result: row.Result, ResultName: auditlabel.ResultName(row.Result), Message: row.Message,
			CreatedAt: convert.TimeToString(row.CreatedAt),
		})
	}
	return &types.AuditOperPageResp{
		Items: items, Page: convert.Int64ToString(int64(page)), PageSize: convert.Int64ToString(int64(pageSize)),
		Total: convert.Int64ToString(total),
	}, nil
}
