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

// FilePutPartHandler 原始分片上传：与 S3 直传一样，响应带 ETag 供 complete 无感回传。
func FilePutPartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 解析定位参数与声明大小/哈希
		q := r.URL.Query()
		key := q.Get("key")
		uploadID := q.Get("upload_id")
		partNumber, err := strconv.Atoi(q.Get("part_number"))
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, errorx.New(400, "part_number 无效"))
			return
		}
		chunkSHA256 := r.Header.Get("X-Chunk-SHA256")
		size := r.ContentLength
		if size < 0 {
			httpx.ErrorCtx(r.Context(), w, errorx.New(400, "必须提供 Content-Length"))
			return
		}

		// 2. 写入分片
		body := io.LimitReader(r.Body, size+1)
		l := filelogic.NewFilePutPartLogic(r.Context(), svcCtx)
		resp, err := l.FilePutPart(key, uploadID, partNumber, size, chunkSHA256, body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		// 3. 与 S3 对齐：PUT 成功后带 ETag（取分片 sha256），前端 complete 统一回传
		if resp != nil && resp.Sha256 != "" {
			w.Header().Set("ETag", `"`+resp.Sha256+`"`)
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
