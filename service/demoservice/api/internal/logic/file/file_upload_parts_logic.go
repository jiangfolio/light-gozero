// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package file

import (
	"context"
	"strconv"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileUploadPartsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadPartsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadPartsLogic {
	return &FileUploadPartsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FileUploadParts 列出已上传分片，供断点续传。
func (l *FileUploadPartsLogic) FileUploadParts(req *types.FileUploadPartsReq) (resp *types.FileUploadPartsResp, err error) {
	// 1. 入参校验
	if req.Key == "" || req.UploadId == "" {
		return nil, errorx.New(400, "key 与 upload_id 不能为空")
	}
	// 2. 拉取已确认分片
	parts, err := l.svcCtx.FileStore.ListParts(l.ctx, req.Key, req.UploadId)
	if err != nil {
		return nil, MapStorageErr(err)
	}
	// 3. 数值字段转字符串
	items := make([]types.FileUploadPartItem, 0, len(parts))
	for _, p := range parts {
		items = append(items, types.FileUploadPartItem{
			PartNumber: strconv.Itoa(p.PartNumber),
			Size:       strconv.FormatInt(p.Size, 10),
			Sha256:     p.SHA256,
		})
	}
	return &types.FileUploadPartsResp{Items: items}, nil
}
