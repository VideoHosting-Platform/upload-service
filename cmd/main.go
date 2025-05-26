package main

import (
	"flag"

	"github.com/VideoHosting-Platform/upload-service/internal/app"
)

func main() {
	configPath := flag.String("config", "configs/dev.yaml", "specify config path")
	flag.Parse()
	app.Run(*configPath)
}
