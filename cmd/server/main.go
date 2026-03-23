package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	sig "github.com/metnagaoglu/holepunch/signal"
)

func main() {
	cfg, err := sig.LoadConfigFromEnv()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	sig.InitLogger(cfg.LogLevel, cfg.LogFormat)

	srv, err := sig.NewServer(cfg)
	if err != nil {
		slog.Error("Failed to create server", "error", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	select {
	case s := <-sigChan:
		slog.Info("Received signal, shutting down", "signal", s)
		srv.Shutdown()
	case err := <-errChan:
		if err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}
}
