package main

import (
	"log"
	"os"

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

	// Create repository based on configuration
	repo := repositories.CreateRepository(cfg)

	// Create handler context with injected dependencies
	handlerCtx := handlers.NewHandlerContext(repo)

	// Initialize server and routes
	srv := server.NewUDPServer(cfg, handlerCtx).
		SetRoutes(router.InitializeRoutes(handlerCtx))

	boundServer, err := srv.Bind()
	if err != nil {
		log.Printf("Failed to bind server: %v", err)
		os.Exit(1)
	}

	if err := boundServer.Listen(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
