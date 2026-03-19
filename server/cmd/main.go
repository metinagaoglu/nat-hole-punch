package main

import (
	"log"
	"os"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/router"
	"udp-hole-punch/pkg/server"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Printf("Invalid configuration: %v", err)
		os.Exit(1)
	}

	log.Printf("Starting UDP Hole Punch Server on %s:%d", cfg.ServerHost, cfg.ServerPort)

	// Initialize server with configuration
	srv := server.NewUDPServer(cfg).SetRoutes(router.InitializeRoutes())

	// Bind to the configured address
	boundServer, err := srv.Bind()
	if err != nil {
		log.Printf("Failed to bind server: %v", err)
		os.Exit(1)
	}

	// Start listening
	if err := boundServer.Listen(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
