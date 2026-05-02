package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Configure installs the process-wide structured logger and returns it.
func Configure(level string, format string) *slog.Logger {
	logger := New(os.Stdout, level, format)
	slog.SetDefault(logger)
	return logger
}

// New creates a slog logger from simple runtime settings.
func New(w io.Writer, level string, format string) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}

	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return slog.New(slog.NewJSONHandler(w, opts))
	default:
		return slog.New(slog.NewTextHandler(w, opts))
	}
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
