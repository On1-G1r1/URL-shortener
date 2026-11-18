package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shortener/internal/config"
	"url-shortener/internal/handler"
	"url-shortener/internal/service"
	"url-shortener/internal/storage/postgres"
	"url-shortener/internal/storage/redis"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pgStore, err := postgres.New(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Error("failed to init postgres", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pgStore.Close()

	redisCache, err := redis.New(ctx, cfg.RedisAddr, cfg.RedisPass, cfg.RedisTTL)
	if err != nil {
		log.Warn("failed to init redis, running without cache", slog.String("err", err.Error()))
		redisCache = nil
	} else {
		defer redisCache.Close()
	}

	shortenerSvc := service.NewShortener(pgStore, redisCache, log)
	httpHandler := handler.New(shortenerSvc, cfg.BaseURL, log)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      httpHandler.InitRoutes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("starting server", slog.String("port", cfg.HTTPPort))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("err", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("forced shutdown", slog.String("err", err.Error()))
	}
	log.Info("server stopped")
}
