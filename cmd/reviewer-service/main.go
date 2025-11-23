package main

import (
	"context"
	"log/slog"
	"net/http"
	"reviewer-service/internal/config"
	"reviewer-service/internal/http-server/handlers/pullrequest"
	"reviewer-service/internal/http-server/handlers/team"
	"reviewer-service/internal/http-server/handlers/user"
	"reviewer-service/internal/http-server/middleware/logger"
	logUtil "reviewer-service/internal/lib/logger/slog"
	"reviewer-service/internal/storage/postgresql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	appConfig := config.MustLoadConfig()
	log := config.MustConfigureLogger(appConfig.Env)

	ctx, cancel := context.WithTimeout(context.Background(), appConfig.Datasource.Timeout)
	defer cancel()

	storage, err := postgresql.NewStorage(&appConfig.Datasource, ctx)
	if err != nil {
		log.Error("Failed to connect to database", logUtil.Err(err))
	}
	defer storage.Close()

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post(
		"/team/add", team.Save(log, storage, storage),
	)

	router.Get(
		"/team/get", team.Get(log, storage),
	)

	router.Post(
		"/users/setIsActive", user.SetIsActive(log, storage, storage),
	)

	router.Get(
		"/users/getReview", user.GetReview(log, storage),
	)

	router.Post(
		"/pullRequest/create", pullrequest.Create(log, storage, storage),
	)

	router.Post(
		"/pullRequest/merge", pullrequest.Merge(log, storage, storage),
	)

	router.Post(
		"/pullRequest/reassign", pullrequest.Reassign(log, storage, storage),
	)

	log.Info("starting service", slog.String("host", appConfig.HttpServer.Host))

	server := &http.Server{
		Addr:         appConfig.HttpServer.Host + ":" + appConfig.HttpServer.Port,
		Handler:      router,
		ReadTimeout:  appConfig.HttpServer.Timeout,
		WriteTimeout: appConfig.HttpServer.Timeout,
		IdleTimeout:  appConfig.HttpServer.IdleTimeout,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Error("server error", logUtil.Err(err))
	}

}
