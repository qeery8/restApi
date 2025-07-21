package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/qeery8/restApi/internal/config"
	"github.com/qeery8/restApi/internal/config/storage/sqlite"
	"github.com/qeery8/restApi/internal/http-server/handlers/url/save"
	"github.com/qeery8/restApi/internal/lib/logger/sl"
	_ "modernc.org/sqlite"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting ...", slog.String("env", cfg.Env))
	log.Debug("debug message are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("falied to init storage", sl.Err(err))
		os.Exit(1) // or return
	}
	_ = storage

	router := chi.NewRouter()

	router.Use(middleware.RequestID) // добавляет request_id в каждый запрос для трейсинга
	router.Use(middleware.Logger)    // логирование
	router.Use(middleware.Recoverer) // если произойдет паника, приложение не должно упасть
	router.Use(middleware.URLFormat) // парсер URL поступающих запросов

	router.Post("/url", save.New(log, storage))

	log.Info("starting server", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("falied to start server")
	}
	log.Error("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
