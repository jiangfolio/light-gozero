package file

import (
	"context"
	"io"
	"strconv"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FilePutPartLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFilePutPartLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FilePutPartLogic {
	return &FilePutPartLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FilePutPart 接收分片二进制流并边写边校验 SHA-256。
func (l *FilePutPartLogic) FilePutPart(key, uploadID string, partNumber int, size int64, chunkSHA256 string, body io.Reader) (*types.FilePutPartResp, error) {
	// 1. 入参校验
	if key == "" || uploadID == "" || partNumber < 1 || size <= 0 || chunkSHA256 == "" {
		return nil, errorx.New(400, "分片参数不完整")
	}
	store := l.svcCtx.FileStore
	if !store.Capabilities().ServerPutPart {
		return nil, errorx.New(400, "当前存储不支持服务端收流分片")
	}

	// 2. 写入分片
	res, err := store.PutPart(l.ctx, key, uploadID, partNumber, body, size, chunkSHA256)
	if err != nil {
		return nil, MapStorageErr(err)
	}

	// 3. 返回分片结果
	return &types.FilePutPartResp{
		PartNumber: strconv.Itoa(res.PartNumber),
		Size:       strconv.FormatInt(res.Size, 10),
		Sha256:     res.SHA256,
	}, nil
}
