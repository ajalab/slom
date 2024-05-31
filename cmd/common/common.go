package common

import (
	"io"
	"log/slog"
)

type CommonFlags struct {
	Debug bool
}

func NewLogger(debug bool, w io.Writer) *slog.Logger {
	var level slog.Level
	if debug {
		level = slog.LevelDebug
	} else {
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)

	return logger
}
