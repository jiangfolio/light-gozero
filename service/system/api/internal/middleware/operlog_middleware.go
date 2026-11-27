package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jiangbyte/light-gozero/common/authctx"
	"github.com/jiangbyte/light-gozero/common/operlog"
	"github.com/jiangbyte/light-gozero/common/reqmeta"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/audit"
)

// OperLogWriter 异步/同步落库操作审计。
type OperLogWriter func(row *audit.OperLog)

type OperLogMiddleware struct {
	write OperLogWriter
	idGen func() (string, error)
}

func NewOperLogMiddleware(write OperLogWriter, idGen func() (string, error)) *OperLogMiddleware {
	return &OperLogMiddleware{write: write, idGen: idGen}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Handle 记录写操作审计；GET 跳过以控制日志量。
func (m *OperLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		start := time.Now()
		bodyPreview := readBodyPreview(r)
		ctx := operlog.WithHolder(r.Context())
		r = r.WithContext(ctx)
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next(rec, r)

		sess, _ := authctx.SessionFrom(r.Context())
		module, action := splitPath(r.URL.Path)
		result := "success"
		if rec.status >= 400 {
			result = "fail"
		}
		id := ""
		if m.idGen != nil {
			id, _ = m.idGen()
		}
		if m.write != nil {
			m.write(&audit.OperLog{
				Id:        id,
				UserId:    sess.UserId,
				Username:  sess.Username,
				Module:    module,
				Action:    action,
				Method:    r.Method,
				Path:      r.URL.Path,
				ReqBody:   bodyPreview,
				Ip:        reqmeta.ClientIP(r.Context()),
				LatencyMs: time.Since(start).Milliseconds(),
				Result:    result,
				Message:   operlog.GetSummary(r.Context()),
			})
		}
	}
}

const (
	// 与 Rest MaxBytes 对齐，保证后续 handler 能拿到完整 body
	maxBodyCapture = 1 << 20
	previewBytes   = 4096
)

func readBodyPreview(r *http.Request) string {
	if r.Body == nil {
		return ""
	}
	// 1. 读完整请求体（有上限），再原样回填，避免大 body 被截断后 Parse 失败
	raw, err := io.ReadAll(io.LimitReader(r.Body, maxBodyCapture))
	if err != nil {
		return ""
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(raw))
	if len(raw) == 0 {
		return ""
	}
	// 2. 审计预览只用前缀，脱敏后截断
	preview := raw
	if len(preview) > previewBytes {
		preview = preview[:previewBytes]
	}
	var m map[string]any
	if json.Unmarshal(preview, &m) == nil {
		for k := range m {
			if sensitiveJSONKey(k) {
				m[k] = "***"
			}
		}
		b, err := json.Marshal(m)
		if err == nil {
			preview = b
		}
	} else if looksSensitiveRaw(preview) {
		return "***"
	}
	s := string(preview)
	if utf8.RuneCountInString(s) > 1800 {
		runes := []rune(s)
		s = string(runes[:1800])
	}
	return s
}

func sensitiveJSONKey(k string) bool {
	lk := strings.ToLower(k)
	keys := []string{"password", "passwd", "token", "secret", "authorization", "private_key", "cipher"}
	for _, s := range keys {
		if strings.Contains(lk, s) {
			return true
		}
	}
	return false
}

func looksSensitiveRaw(raw []byte) bool {
	low := strings.ToLower(string(raw))
	return strings.Contains(low, "password") || strings.Contains(low, "private_key") || strings.Contains(low, "token")
}

func splitPath(path string) (module, action string) {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[2], parts[3]
	}
	if len(parts) >= 2 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	return "", path
}
