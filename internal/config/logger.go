package config

import (
	"log/slog"
	"os"
)

const (
	devEnv  = "dev"
	testEnv = "test"
	prodEnv = "prod"
)

func MustConfigureLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case devEnv:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case testEnv:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case prodEnv:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return logger
}
