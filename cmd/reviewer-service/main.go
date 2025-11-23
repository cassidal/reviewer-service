package main

import (
	"os"
	"reviewer-service/internal/config"
	"reviewer-service/internal/lib/logger/slog"
	"reviewer-service/internal/storage/postgresql"
)

func main() {
	appConfig := config.MustLoadConfig()
	log := config.MustConfigureLogger(appConfig.Env)

	storage, err := postgresql.New(appConfig.Datasource)
	if err != nil {
		log.Error("failed to initialize storage", slog.Err(err))
		os.Exit(1)
	}

	_ = storage

	log.Info("starting reviewer service")
}
