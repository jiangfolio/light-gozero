// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"net/url"
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

// FileUploadPartUrl 统一下发分片 PUT 地址：local 为同域 /file/upload/part，s3 为预签名 URL。
func (l *FileUploadPartUrlLogic) FileUploadPartUrl(req *types.FileUploadPartUrlReq) (resp *types.FileUploadPartUrlResp, err error) {
	// 1. 入参校验
	if req.Key == "" || req.UploadId == "" || req.PartNumber == "" {
		return nil, errorx.New(400, "key/upload_id/part_number 不能为空")
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
	store := l.svcCtx.FileStore

	// 2. local：校验会话后返回同域相对 URL，前端无感切换
	if store.Capabilities().ServerPutPart {
		if _, prepErr := store.PreparePart(l.ctx, req.Key, req.UploadId, partNumber, storage.PartOpts{
			ChunkSHA256: req.ChunkSha256,
			Size:        size,
		}); prepErr != nil {
			return nil, MapStorageErr(prepErr)
		}
		q := url.Values{}
		q.Set("key", req.Key)
		q.Set("upload_id", req.UploadId)
		q.Set("part_number", strconv.Itoa(partNumber))
		headers := map[string]string{
			"Content-Type": "application/octet-stream",
		}
		if req.ChunkSha256 != "" {
			headers["X-Chunk-SHA256"] = req.ChunkSha256
		}
		expire := time.Now().UTC().Add(time.Duration(l.svcCtx.Config.Storage.UploadExpireSeconds) * time.Second)
		if l.svcCtx.Config.Storage.UploadExpireSeconds <= 0 {
			expire = time.Now().UTC().Add(24 * time.Hour)
		}
		return &types.FileUploadPartUrlResp{
			UploadUrl: "/api/v1/file/upload/part?" + q.Encode(),
			Headers:   headers,
			ExpiresAt: expire.Format(time.RFC3339),
		}, nil
	}

	// 3. s3：预签名直传 URL
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
