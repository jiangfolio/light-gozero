package file

import (
	"io"
	"net/http"
	"strconv"

	"github.com/jiangbyte/light-gozero/common/errorx"
	filelogic "github.com/jiangbyte/light-gozero/service/demoservice/api/internal/logic/file"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// FileDownloadHandler 经应用流式下载（local Open / s3 GetObject），demo 无鉴权。
func FileDownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			httpx.ErrorCtx(r.Context(), w, errorx.New(400, "key 不能为空"))
			return
		}
		rc, info, err := svcCtx.FileStore.Open(r.Context(), key)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, filelogic.MapStorageErr(err))
			return
		}
		defer rc.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Disposition", "attachment")
		if info.Size > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
		}
		if info.Integrity.MerkleSHA256 != "" {
			w.Header().Set("X-Content-Merkle-SHA256", info.Integrity.MerkleSHA256)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, rc)
	}
}
