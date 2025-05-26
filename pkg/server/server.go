package server

import (
	"context"
	"net/http"
)

type Server struct {
	httpServer http.Server
}

type Config struct {
	Port string `yaml:"port"`
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
