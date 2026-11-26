package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNotFoundHandlerUnifiedBody(t *testing.T) {
	// 1. 模拟未匹配路由，断言 JSON 包络与业务码
	req := httptest.NewRequest(http.MethodGet, "/not-exist", nil)
	rec := httptest.NewRecorder()
	NotFoundHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Code != codeNotFound || body.Message != msgNotFound {
		t.Fatalf("body = %+v", body)
	}
}

func TestNotAllowedHandlerUnifiedBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/any", nil)
	rec := httptest.NewRecorder()
	NotAllowedHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	var body Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Code != codeNotAllowed || body.Message != msgNotAllowed {
		t.Fatalf("body = %+v", body)
	}
}
