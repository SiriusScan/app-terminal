package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/SiriusScan/go-api/sirius/queue"
)

type Command struct {
	Command       string `json:"command"`
	UserID        string `json:"userId"`
	Timestamp     string `json:"timestamp"`
	ResponseQueue string `json:"responseQueue"`
}

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	ps     *PowerShell
	mu     sync.Mutex // Protects PowerShell instance
}

func NewManager() (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize PowerShell
	ps, err := NewPowerShell()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize PowerShell: %v", err)
	}

	return &Manager{
		ctx:    ctx,
		cancel: cancel,
		ps:     ps,
	}, nil
}

func (m *Manager) ListenForCommands() {
	log.Println("Starting command listener...")
	queue.Listen("terminal", m.handleCommand)
}

func (m *Manager) handleCommand(msg string) {
	var cmd Command
	if err := json.Unmarshal([]byte(msg), &cmd); err != nil {
		log.Printf("Failed to parse command: %v", err)
		return
	}

	// Skip empty commands (used for connection test)
	if cmd.Command == "" {
		if cmd.ResponseQueue != "" {
			if err := queue.Send(cmd.ResponseQueue, "Connected"); err != nil {
				log.Printf("Failed to send connection confirmation: %v", err)
			}
		}
		return
	}

	response := m.executeCommand(cmd)

	if cmd.ResponseQueue != "" {
		if err := queue.Send(cmd.ResponseQueue, response); err != nil {
			log.Printf("Failed to send response: %v", err)
		}
	}
}

func (m *Manager) executeCommand(cmd Command) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ps == nil {
		return "Error: PowerShell session not initialized"
	}

	// Format table output commands
	command := cmd.Command
	if strings.HasPrefix(command, "Get-") || strings.Contains(command, "| Format-Table") {
		if !strings.Contains(command, "| Format-Table") && !strings.Contains(command, "|ft") {
			command += " | Format-Table -AutoSize | Out-String -Width 120"
		}
	}

	output, err := m.ps.Execute(command)
	if err != nil {
		log.Printf("Failed to execute command: %v", err)
		return fmt.Sprintf("Error executing command: %v", err)
	}

	return output
}

func (m *Manager) Shutdown() {
	log.Println("Shutting down terminal manager...")
	m.cancel()
}
