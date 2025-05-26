package config

import (
	"fmt"

	"github.com/VideoHosting-Platform/upload-service/pkg/server"
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string        `yaml:"env"`
	HTTP server.Config `yaml:"http"`
}

func MustLoad(configPath string) *Config {

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Sprintf("error occured while reading config: %s", err.Error()))
	}
	if cfg.Env != "dev" && cfg.Env != "test" && cfg.Env != "prod" {
		panic("error occured while reading config - not valid env value")
	}
	return &cfg
}
