package main

import (
	"flag"
	"log"
	"os"

	"send-data/internal/config"
	"send-data/internal/server"
)

func main() {
	configPath := flag.String("config", "config/send-data.yaml", "path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := os.MkdirAll(cfg.Storage.SpoolDir, 0700); err != nil {
		log.Fatalf("spool dir: %v", err)
	}

	srv := server.New(cfg)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server: %v", err)
	}
}
