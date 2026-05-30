package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// Setup configures the slog default logger based on the given level and
// optional file path. If file is empty, logs go to stderr.
// Returns a close function to cleanly shut down the logger (close log file).
func Setup(levelStr, file string) (func(), error) {
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
	var closer func()
	if file != "" {
		f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, err
		}
		writer = f
		closer = func() { f.Close() }
	} else {
		closer = func() {} // no-op for stderr
	}

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
	return closer, nil
}
