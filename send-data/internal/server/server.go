package server

import (
	"fmt"
	"log"
	"net/http"

	"send-data/internal/config"
	"send-data/internal/handler"
)

type Server struct {
	cfg *config.Config
	srv *http.Server
}

func New(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	handler.New(cfg).Register(mux)

	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:    cfg.Addr(),
			Handler: mux,
		},
	}
}

func (s *Server) ListenAndServe() error {
	log.Printf("send-data listening on %s (spool=%s)", s.cfg.Addr(), s.cfg.Storage.SpoolDir)

	if s.cfg.Server.TLS.Enabled {
		return fmt.Errorf("tls is not implemented yet; set server.tls.enabled=false")
	}

	return s.srv.ListenAndServe()
}
