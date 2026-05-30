package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Setup configures the slog default logger based on the given level and
// optional file path. If file is empty, logs go to stderr.
func Setup(levelStr, file string) error {
	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var writer io.Writer = os.Stderr
	if file != "" {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		writer = f
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
	return nil
}
