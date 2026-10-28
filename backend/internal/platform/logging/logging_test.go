package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDiagnosticsStayOnOneJSONLine(t *testing.T) {
	var out bytes.Buffer
	logger := New(&out, "web", "test", "fixture", "INFO")
	Relay(logger, nil, "nginx diagnostic: \"quoted\"\nsecond line\t")
	Relay(logger, nil, `connection refused, client: 127.0.0.1, request: "GET /health?token=never-log-this"`)
	Relay(logger, map[string]any{"status": 503, "request_id": "abc", "duration_ms": 1}, "")
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines", len(lines))
	}
	if strings.Contains(out.String(), "never-log-this") {
		t.Fatal("query credential leaked")
	}
	for _, line := range lines {
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"timestamp", "level", "service", "env", "commit", "event"} {
			if _, ok := row[key]; !ok {
				t.Fatalf("missing %s", key)
			}
		}
	}
}
