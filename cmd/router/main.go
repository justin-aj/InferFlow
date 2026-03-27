package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ajinfrank/inferflow/internal/server"
)

func main() {
	cfg, err := server.LoadConfigFromEnv()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("create server: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop
		if err := srv.Shutdown(); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("inferflow router listening on %s", cfg.ListenAddr)
	if err := srv.Run(); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
