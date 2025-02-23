package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/SiriusScan/app-terminal/internal/terminal"
)

func main() {
	fmt.Println("Terminal service is running...")

	// Create and initialize terminal manager
	manager, err := terminal.NewManager()
	if err != nil {
		log.Fatalf("Failed to create terminal manager: %v", err)
	}

	// Start listening for terminal commands
	manager.ListenForCommands()
	defer manager.Shutdown()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down terminal service...")
}
