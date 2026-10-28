package httpapi

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"traceable-ai-quiz/internal/application/health"
)

//go:embed openapi.json
var contract embed.FS

var PublicRoutes = map[string]string{"/health": "get", "/api/health": "get", "/api/openapi.json": "get"}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

func New(service health.Service, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		var id [16]byte
		_, _ = rand.Read(id[:])
		requestID := hex.EncodeToString(id[:])
		out := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		out.Header().Set("Content-Type", "application/json; charset=utf-8")
		out.Header().Set("Cache-Control", "no-store")
		out.Header().Set("X-Request-ID", requestID)
		defer func() {
			log.Info("http_request", "request_id", requestID, "method", r.Method, "path", r.URL.Path, "status", out.status, "duration_ms", time.Since(started).Milliseconds())
		}()
		fail := func(status int, code, message string) {
			out.WriteHeader(status)
			_ = json.NewEncoder(out).Encode(apiError{code, message, requestID})
		}
		if _, ok := PublicRoutes[r.URL.Path]; !ok {
			if r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/") {
				fail(401, "UNAUTHENTICATED", "此业务入口尚未启用")
			} else {
				fail(404, "NOT_FOUND", "资源不存在")
			}
			return
		}
		if r.Method != http.MethodGet {
			out.Header().Set("Allow", "GET")
			fail(405, "METHOD_NOT_ALLOWED", "仅支持 GET")
			return
		}
		if r.URL.RawQuery != "" {
			fail(422, "VALIDATION_ERROR", "此接口不接受查询参数")
			return
		}
		if r.URL.Path == "/api/openapi.json" {
			data, _ := contract.ReadFile("openapi.json")
			_, _ = out.Write(data)
			return
		}
		_ = json.NewEncoder(out).Encode(service.Status())
	})
}
