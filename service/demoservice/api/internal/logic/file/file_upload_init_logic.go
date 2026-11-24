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

type FileUploadInitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileUploadInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileUploadInitLogic {
	return &FileUploadInitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FileUploadInit 创建分片上传会话；上传路径统一为 part_url。
func (l *FileUploadInitLogic) FileUploadInit(req *types.FileUploadInitReq) (resp *types.FileUploadInitResp, err error) {
	// 1. 校验 key
	if req.Key == "" {
		return nil, errorx.New(400, "key 不能为空")
	}
	store := l.svcCtx.FileStore

	// 2. 发起上传会话
	uploadID, err := store.Initiate(l.ctx, req.Key, storage.InitiateOpts{
		ContentType: req.ContentType,
	})
	if err != nil {
		return nil, MapStorageErr(err)
	}

	// 3. 组装响应：UploadVia 固定 part_url，前端不必按 provider 分支
	resp = &types.FileUploadInitResp{
		UploadId:   uploadID,
		Provider:   store.Provider(),
		AccessMode: store.AccessMode(),
		ChunkSize:  strconv.FormatInt(store.ChunkSize(), 10),
		MaxParts:   strconv.Itoa(store.MaxParts()),
		UploadVia:  "part_url",
	}
	if store.AccessMode() == storage.AccessPublic {
		if u, uErr := store.PublicURL(req.Key); uErr == nil {
			resp.PublicUrl = u
		}
	}
	return resp, nil
}
