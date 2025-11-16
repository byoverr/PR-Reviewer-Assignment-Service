package logger

import (
	"log/slog"
	"os"
)

func New(level, output, filePath string) *slog.Logger {
	var h slog.Handler
	var lvl slog.Level

	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	if output == "file" {
		//nolint:gosec
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			h = slog.NewTextHandler(os.Stdout, opts)
		} else {
			h = slog.NewTextHandler(f, opts)
		}
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(h)
}
