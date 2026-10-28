package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"traceable-ai-quiz/internal/application/health"
	"traceable-ai-quiz/internal/platform/logging"
)

func TestNormalFailureBoundaryRoutes(t *testing.T) {
	var logs bytes.Buffer
	handler := New(health.New("fixture", "test"), logging.New(&logs, "api", "test", "fixture", "INFO"))
	for _, tc := range []struct {
		method, path, code string
		status             int
	}{
		{"GET", "/health", "", 200}, {"GET", "/api/health", "", 200},
		{"GET", "/api/openapi.json", "", 200}, {"POST", "/health", "METHOD_NOT_ALLOWED", 405},
		{"GET", "/missing-resource", "NOT_FOUND", 404}, {"GET", "/api/classes/C1/quiz/runs", "UNAUTHENTICATED", 401},
		{"GET", "/health?token=never-log-this", "VALIDATION_ERROR", 422},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer never-log-this")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != tc.status {
			t.Fatalf("%s got %d", tc.path, rec.Code)
		}
		if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") || rec.Header().Get("X-Request-ID") == "" {
			t.Fatal("missing JSON or request ID")
		}
		var result map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if tc.code != "" && result["code"] != tc.code {
			t.Fatalf("%s wrong code", tc.path)
		}
	}
	if strings.Contains(logs.String(), "never-log-this") {
		t.Fatal("credential leaked into logs")
	}
}

func TestContractAndRoutesBidirectional(t *testing.T) {
	data, err := contract.ReadFile("openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		OpenAPI string                                `json:"openapi"`
		Paths   map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.OpenAPI != "3.1.0" || len(document.Paths) != len(PublicRoutes) {
		t.Fatal("contract route drift")
	}
	for path, method := range PublicRoutes {
		raw, ok := document.Paths[path][method]
		if !ok {
			t.Fatalf("missing route %s", path)
		}
		var op map[string]json.RawMessage
		_ = json.Unmarshal(raw, &op)
		var security []any
		if value, ok := op["security"]; !ok {
			t.Fatalf("public route %s missing security", path)
		} else if err := json.Unmarshal(value, &security); err != nil || len(security) != 0 {
			t.Fatal("public security drift")
		}
		for key := range document.Paths[path] {
			if key != method {
				t.Fatalf("undeclared operation %s %s", key, path)
			}
		}
	}
}
