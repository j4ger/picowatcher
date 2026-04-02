package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/j4ger/picowatcher/config"
	"gopkg.in/lumberjack.v2"
)

func Setup(cfg config.LogConfig) *slog.Logger {
	roller := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
	w := io.MultiWriter(os.Stdout, roller)
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}
