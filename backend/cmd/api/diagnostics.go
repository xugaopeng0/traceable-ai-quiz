package main

import (
	"log"
	"log/slog"
	"strings"
)

type diagnosticWriter struct{ log *slog.Logger }

func (w diagnosticWriter) Write(p []byte) (int, error) {
	w.log.Error("http_diagnostic", "message", strings.TrimSpace(string(p)))
	return len(p), nil
}
func slogAdapter(logger *slog.Logger) *log.Logger { return log.New(diagnosticWriter{logger}, "", 0) }
