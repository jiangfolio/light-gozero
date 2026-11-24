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

// FileUploadInit 创建分片上传会话，返回分片大小与可选公开直链前缀信息。
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

	// 3. 组装响应；public 模式下附带稳定公开 URL（文件尚未 complete 前仅作预览地址）
	resp = &types.FileUploadInitResp{
		UploadId:   uploadID,
		Provider:   store.Provider(),
		AccessMode: store.AccessMode(),
		ChunkSize:  strconv.FormatInt(store.ChunkSize(), 10),
		MaxParts:   strconv.Itoa(store.MaxParts()),
	}
	if store.AccessMode() == storage.AccessPublic {
		if u, uErr := store.PublicURL(req.Key); uErr == nil {
			resp.PublicUrl = u
		}
	}
	return resp, nil
}
