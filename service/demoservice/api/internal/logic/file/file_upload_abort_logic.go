// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUploadAbortLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadAbortLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadAbortLogic {
	return &FileUploadAbortLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FileUploadAbort 中止上传并清理临时分片。
func (l *FileUploadAbortLogic) FileUploadAbort(req *types.FileUploadAbortReq) error {
	// 1. 入参校验
	if req.Key == "" || req.UploadId == "" {
		return errorx.New(400, "key 与 upload_id 不能为空")
	}
	// 2. 中止会话（幂等）
	if err := l.svcCtx.FileStore.Abort(l.ctx, req.Key, req.UploadId); err != nil {
		return MapStorageErr(err)
	}
	return nil
}
