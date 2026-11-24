package file

import (
	"io"
	"net/http"
	"strconv"

	"github.com/jiangbyte/light-gozero/common/errorx"
	"github.com/jiangbyte/light-gozero/common/storage"
	filelogic "github.com/jiangbyte/light-gozero/service/demoservice/api/internal/logic/file"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// FilePublicHandler 稳定公开直链下载（无 TTL），仅 AccessMode=public 时可用。
func FilePublicHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 模式与 key 校验
		if svcCtx.FileStore.AccessMode() != storage.AccessPublic {
			httpx.ErrorCtx(r.Context(), w, errorx.New(403, "当前非公开访问模式"))
			return
		}
		key := r.URL.Query().Get("key")
		if key == "" {
			httpx.ErrorCtx(r.Context(), w, errorx.New(400, "key 不能为空"))
			return
		}

		// 2. 打开对象并流式写出（不走 JSON 包络）
		rc, info, err := svcCtx.FileStore.Open(r.Context(), key)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, filelogic.MapStorageErr(err))
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", "attachment")
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
		if info.Integrity.MerkleSHA256 != "" {
			w.Header().Set("X-Content-Merkle-SHA256", info.Integrity.MerkleSHA256)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, rc)
	}
}
