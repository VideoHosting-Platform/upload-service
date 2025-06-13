package server

import (
	"context"
	"net/http"
)

type Server struct {
	httpServer http.Server
}

type Config struct {
	Port string `env:"HTTP_PORT" env-default:"8080"`
}

func NewServer(cfg *Config, handler http.Handler) *Server {
	return &Server{
		httpServer: http.Server{
			Addr:    ":" + cfg.Port,
			Handler: handler,
		},
	}
}

func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
