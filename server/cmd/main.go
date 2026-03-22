package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/handlers"
	"udp-hole-punch/pkg/repositories"
	"udp-hole-punch/pkg/router"
	"udp-hole-punch/pkg/server"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		log.Printf("Invalid configuration: %v", err)
		os.Exit(1)
	}

	log.Printf("Starting UDP Hole Punch Server on %s:%d", cfg.ServerHost, cfg.ServerPort)

	repo := repositories.CreateRepository(cfg)
	handlerCtx := handlers.NewHandlerContext(repo, cfg.ClientTTL)

	srv := server.NewUDPServer(cfg, handlerCtx).
		SetRoutes(router.InitializeRoutes(handlerCtx))

	boundServer, err := srv.Bind()
	if err != nil {
		log.Printf("Failed to bind server: %v", err)
		os.Exit(1)
	}

	// Graceful shutdown on SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- boundServer.Listen()
	}()

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v, shutting down...", sig)
		boundServer.Shutdown()
	case err := <-errChan:
		if err != nil {
			log.Printf("Server error: %v", err)
			os.Exit(1)
		}
	}
}
