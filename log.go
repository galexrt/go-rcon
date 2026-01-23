package rcon

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})
	logger = slog.New(handler)
}

// SetLog set logr logger
func SetLog(l *slog.Logger) {
	logger = l
}
