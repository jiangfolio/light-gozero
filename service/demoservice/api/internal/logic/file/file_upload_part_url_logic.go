// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"strconv"
	"time"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/storage"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUploadPartUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadPartUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadPartUrlLogic {
	return &FileUploadPartUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FileUploadPartUrl 下发 S3 分片预签名 URL（local 模式不适用）。
func (l *FileUploadPartUrlLogic) FileUploadPartUrl(req *types.FileUploadPartUrlReq) (resp *types.FileUploadPartUrlResp, err error) {
	// 1. 入参与能力校验
	if req.Key == "" || req.UploadId == "" || req.PartNumber == "" {
		return nil, errorx.New(400, "key/upload_id/part_number 不能为空")
	}
	store := l.svcCtx.FileStore
	if !store.Capabilities().DirectUpload {
		return nil, errorx.New(400, "当前存储不支持预签名直传")
	}
	partNumber, err := strconv.Atoi(req.PartNumber)
	if err != nil || partNumber < 1 {
		return nil, errorx.New(400, "part_number 无效")
	}
	var size int64
	if req.Size != "" {
		size, err = strconv.ParseInt(req.Size, 10, 64)
		if err != nil || size < 0 {
			return nil, errorx.New(400, "size 无效")
		}
	}

	// 2. 申请预签名（可选绑定分片 SHA-256）
	prep, err := store.PreparePart(l.ctx, req.Key, req.UploadId, partNumber, storage.PartOpts{
		ChunkSHA256: req.ChunkSha256,
		Size:        size,
	})
	if err != nil {
		return nil, MapStorageErr(err)
	}
	return &types.FileUploadPartUrlResp{
		UploadUrl: prep.UploadURL,
		Headers:   prep.Headers,
		ExpiresAt: prep.ExpiresAt.UTC().Format(time.RFC3339),
	}, nil
}
