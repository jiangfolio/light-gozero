package file

import (
	"errors"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/storage"
)

// MapStorageErr 将存储哨兵错误映射为对外业务错误，避免暴露盘符路径。
func MapStorageErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, storage.ErrInvalidKey):
		return errorx.New(400, "对象键无效")
	case errors.Is(err, storage.ErrInvalidPart):
		return errorx.New(400, "分片参数无效")
	case errors.Is(err, storage.ErrChecksumMismatch), errors.Is(err, storage.ErrMerkleMismatch):
		return errorx.New(400, "校验失败")
	case errors.Is(err, storage.ErrPartExistsConflict):
		return errorx.New(409, "分片已存在且校验不一致")
	case errors.Is(err, storage.ErrUploadNotFound), errors.Is(err, storage.ErrObjectNotFound):
		return errorx.New(404, "资源不存在")
	case errors.Is(err, storage.ErrUploadExpired):
		return errorx.New(410, "上传会话已过期")
	case errors.Is(err, storage.ErrUploadState):
		return errorx.New(409, "上传会话状态不允许此操作")
	case errors.Is(err, storage.ErrTooLarge), errors.Is(err, storage.ErrInsufficientSpace):
		return errorx.New(413, "对象过大或磁盘空间不足")
	case errors.Is(err, storage.ErrDirectUploadOnly):
		return errorx.New(400, "当前存储仅支持直传")
	case errors.Is(err, storage.ErrNotSupported):
		return errorx.New(400, "当前模式不支持该操作")
	default:
		return err
	}
}
