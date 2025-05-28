package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/VideoHosting-Platform/upload-service/internal/handler"
	"github.com/VideoHosting-Platform/upload-service/pkg/config"
	"github.com/VideoHosting-Platform/upload-service/pkg/minio_connection"
	"github.com/VideoHosting-Platform/upload-service/pkg/queue"
	"github.com/VideoHosting-Platform/upload-service/pkg/server"
)

func Run(configPath string) {
	cfg := config.MustLoad(configPath)

	mc, err := minio_connection.NewClient(&cfg.Minio)
	if err != nil {
		fmt.Println("minio client err") // !change
	}

	q, err := queue.New(&cfg.RabbitMQ)
	if err != nil {

	}

	handler := handler.New(
		mc,
		q,
	)
	server := server.NewServer(&cfg.HTTP, handler.Init())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		if err := server.Run(); err != nil {
			fmt.Println("shutdown") // ! change
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		fmt.Println("error while shutdown server") // ! change
	}
}
