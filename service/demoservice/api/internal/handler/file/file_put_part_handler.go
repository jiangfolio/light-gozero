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

// FilePutPartHandler 原始分片上传：query 传 key/upload_id/part_number，header 传校验与长度。
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

		// 2. 限制可读字节，防止超读
		body := io.LimitReader(r.Body, size+1)
		l := filelogic.NewFilePutPartLogic(r.Context(), svcCtx)
		resp, err := l.FilePutPart(key, uploadID, partNumber, size, chunkSHA256, body)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
