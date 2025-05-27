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
	"github.com/VideoHosting-Platform/upload-service/pkg/server"
)

// TODO event posting

func Run(configPath string) {
	cfg := config.MustLoad(configPath)
	fmt.Println(cfg)

	mc, err := minio_connection.NewClient(&cfg.Minio)
	if err != nil {
		fmt.Println("minio client err") // !change
	}

	handler := handler.New(mc, cfg.Minio.BucketName)
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
