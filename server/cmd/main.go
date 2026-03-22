package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"udp-hole-punch/pkg/config"
	"udp-hole-punch/pkg/handlers"
	"udp-hole-punch/pkg/logger"
	"udp-hole-punch/pkg/repositories"
	"udp-hole-punch/pkg/router"
	"udp-hole-punch/pkg/server"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		slog.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logger.Init(cfg.LogLevel, cfg.LogFormat)

	slog.Info("Starting UDP Hole Punch Server", "host", cfg.ServerHost, "port", cfg.ServerPort)

	repo := repositories.CreateRepository(cfg)
	handlerCtx := handlers.NewHandlerContext(repo, cfg.ClientTTL)

	srv := server.NewUDPServer(cfg, handlerCtx).
		SetRoutes(router.InitializeRoutes(handlerCtx))

	boundServer, err := srv.Bind()
	if err != nil {
		slog.Error("Failed to bind server", "error", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- boundServer.Listen()
	}()

	select {
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down", "signal", sig)
		boundServer.Shutdown()
	case err := <-errChan:
		if err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
