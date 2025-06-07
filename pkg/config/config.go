package config

import (
	"fmt"
	"os"

	"github.com/VideoHosting-Platform/upload-service/pkg/minio_connection"
	"github.com/VideoHosting-Platform/upload-service/pkg/queue"
	"github.com/VideoHosting-Platform/upload-service/pkg/server"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string `env:"APP_ENV" env-default:"dev"`
	HTTP     server.Config
	Minio    minio_connection.Config
	RabbitMQ queue.Config
}

func MustLoad() *Config {

	var cfg Config

	envFile := ".env"

	// Проверяем наличие .env файла
	if _, err := os.Stat(envFile); err == nil {
		// Если файл есть, загружаем из него (переменные окружения перезапишут значения из файла)
		err = cleanenv.ReadConfig(envFile, &cfg)
		if err != nil {
			panic(fmt.Sprintf("Failed to read config from .env: %v", err))
		}
	} else {
		// Если файла нет, загружаем только из переменных окружения
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			panic(fmt.Sprintf("Failed to read config from env vars: %v", err))
		}
	}
	return &cfg
}
