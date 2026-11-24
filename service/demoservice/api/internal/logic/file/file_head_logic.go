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

type FileHeadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileHeadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileHeadLogic {
	return &FileHeadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// FileHead 查询已完成对象元信息。
func (l *FileHeadLogic) FileHead(req *types.FileHeadReq) (resp *types.FileHeadResp, err error) {
	// 1. 入参校验
	if req.Key == "" {
		return nil, errorx.New(400, "key 不能为空")
	}
	store := l.svcCtx.FileStore

	// 2. Head 对象
	info, err := store.Head(l.ctx, req.Key)
	if err != nil {
		return nil, MapStorageErr(err)
	}

	// 3. 组装响应
	resp = &types.FileHeadResp{
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
