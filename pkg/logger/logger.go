// Package logger builds the process-wide structured logger.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON logger. JSON is the default because the first consumer of
// these logs is a log pipeline, not a human tailing a terminal.
func New(level string) *slog.Logger {
	var parsed slog.Level
	switch strings.ToUpper(level) {
	case "DEBUG":
		parsed = slog.LevelDebug
	case "WARN":
		parsed = slog.LevelWarn
	case "ERROR":
		parsed = slog.LevelError
	default:
		parsed = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parsed})
	return slog.New(handler)
}
