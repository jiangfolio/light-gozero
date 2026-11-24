// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"strconv"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/storage"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUploadCompleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadCompleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadCompleteLogic {
	return &FileUploadCompleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FileUploadComplete 合并分片；S3 需带 parts(etag)，local 可省略 parts。
func (l *FileUploadCompleteLogic) FileUploadComplete(req *types.FileUploadCompleteReq) (resp *types.FileUploadCompleteResp, err error) {
	// 1. 入参校验
	if req.Key == "" || req.UploadId == "" {
		return nil, errorx.New(400, "key 与 upload_id 不能为空")
	}
	store := l.svcCtx.FileStore

	// 2. 解析客户端回传的分片（S3 直传后必须带 ETag）
	parts := make([]storage.PartResult, 0, len(req.Parts))
	for _, p := range req.Parts {
		n, convErr := strconv.Atoi(p.PartNumber)
		if convErr != nil || n < 1 {
			return nil, errorx.New(400, "parts.part_number 无效")
		}
		var size int64
		if p.Size != "" {
			size, convErr = strconv.ParseInt(p.Size, 10, 64)
			if convErr != nil {
				return nil, errorx.New(400, "parts.size 无效")
			}
		}
		parts = append(parts, storage.PartResult{
			PartNumber: n,
			Size:       size,
			SHA256:     p.Sha256,
			ETag:       p.Etag,
		})
	}

	// 3. Complete
	info, err := store.Complete(l.ctx, req.Key, req.UploadId, parts, req.MerkleSha256)
	if err != nil {
		return nil, MapStorageErr(err)
	}

	// 4. 组装结果
	resp = &types.FileUploadCompleteResp{
		Key:          info.Key,
		Size:         strconv.FormatInt(info.Size, 10),
		MerkleSha256: info.Integrity.MerkleSHA256,
		ChunkCount:   strconv.Itoa(info.Integrity.ChunkCount),
	}
	if store.AccessMode() == storage.AccessPublic {
		if u, uErr := store.PublicURL(info.Key); uErr == nil {
			resp.PublicUrl = u
		}
	}
	return resp, nil
}
