package app

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/VideoHosting-Platform/upload-service/internal/handler"
	"github.com/VideoHosting-Platform/upload-service/pkg/config"
	"github.com/VideoHosting-Platform/upload-service/pkg/logger"
	"github.com/VideoHosting-Platform/upload-service/pkg/minio_connection"
	"github.com/VideoHosting-Platform/upload-service/pkg/queue"
	"github.com/VideoHosting-Platform/upload-service/pkg/server"
)

func Run() {

	cfg := config.MustLoad()

	logger.Init(cfg.Env)
	log := logger.WithSource("app")
	log.Debug("configuration loaded", "cfg", cfg)

	mc, err := minio_connection.NewClient(&cfg.Minio)
	if err != nil {
		log.Error("failed to initialize MinIO client", "error", err)
		os.Exit(1)
	}
	log.Info("MinIO client initialized successfully")

	q, err := queue.New(&cfg.RabbitMQ)
	if err != nil {
		log.Error("failed to initialize RabbitMQ connection", "error", err)
		os.Exit(1)
	}
	log.Info("RabbitMQ connection established")

	handler := handler.New(
		mc,
		q,
	)
	server := server.NewServer(&cfg.HTTP, handler.Init())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := server.Run(); err != nil {
			log.Error("HTTP server stopped with error", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("shutting down HTTP server", "timeout", "10s")
	if err := server.Stop(ctx); err != nil {
		log.Error("error while shutting down server", "error", err)
	}
	if err := q.Close(); err != nil {
		log.Error("error while shutting down queue connection", "error", err)
	}
}
