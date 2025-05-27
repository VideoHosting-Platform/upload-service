package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/VideoHosting-Platform/upload-service/internal/handler"
	"github.com/VideoHosting-Platform/upload-service/pkg/config"
	"github.com/VideoHosting-Platform/upload-service/pkg/minio_connection"
	"github.com/VideoHosting-Platform/upload-service/pkg/server"
	"github.com/VideoHosting-Platform/upload-service/pkg/tokenutil"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Run(configPath string) {
	cfg := config.MustLoad(configPath)
	fmt.Println(cfg)

	tm, err := tokenutil.New(&cfg.JWT)
	if err != nil {
		fmt.Println("token manager err", err.Error()) // ! change
	}

	mc, err := minio_connection.NewClient(&cfg.Minio)
	if err != nil {
		fmt.Println("minio client err") // !change
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err) // ! change
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err) // ! change
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"video_processing", // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		fmt.Println("queue declare error", err.Error()) // ! change
	}

	handler := handler.New(
		tm,
		mc,
		cfg.Minio.BucketName,
		ch,
		q.Name,
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
