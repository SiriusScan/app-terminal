package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/SiriusScan/app-terminal/internal/terminal"
	"github.com/SiriusScan/go-api/sirius/slogger"
)

func main() {
	slogger.Init()

	slog.Info("terminal service starting")

	// Create and initialize terminal manager
	manager, err := terminal.NewManager()
	if err != nil {
		slog.Error("failed to create terminal manager", "error", err)
		os.Exit(1)
	}

	// Start listening for terminal commands
	manager.ListenForCommands()
	defer manager.Shutdown()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down terminal service")
}
