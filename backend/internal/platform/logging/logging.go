package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

func New(out io.Writer, service, env, commit, level string) *slog.Logger {
	var minimum slog.Level
	if err := minimum.UnmarshalText([]byte(level)); err != nil {
		minimum = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{Level: minimum, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			a.Key = "timestamp"
		}
		if a.Key == slog.MessageKey {
			a.Key = "event"
		}
		return a
	}})
	return slog.New(handler).With("service", service, "env", env, "commit", commit)
}

// Relay keeps nginx diagnostics and access output within the same JSON contract.
func Relay(log *slog.Logger, fields map[string]any, raw string) {
	if fields == nil {
		// nginx appends URI/query information after the client context. Keep
		// the diagnostic itself, never credentials that might occur in a URI.
		if index := strings.Index(raw, ", client:"); index >= 0 {
			raw = raw[:index]
		}
		log.Error("nginx_diagnostic", "message", raw)
		return
	}
	attrs := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		if key != "event" {
			attrs = append(attrs, key, value)
		}
	}
	log.Log(context.Background(), slog.LevelInfo, "http_request", attrs...)
}
