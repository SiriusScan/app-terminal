package terminal

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

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
	mu     sync.Mutex     // Protects PowerShell instance
	logger *LoggingClient // Centralized logging client
}

func NewManager() (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize logging client
	logger := NewLoggingClient()
	logger.LogServiceLifecycle("initializing", map[string]interface{}{
		"service": "sirius-terminal",
	})

	// Initialize PowerShell
	ps, err := NewPowerShell()
	if err != nil {
		cancel()
		logger.LogPowerShellInitialization(false, map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize PowerShell: %v", err)
	}

	logger.LogPowerShellInitialization(true, map[string]interface{}{
		"powershell_available": true,
	})

	return &Manager{
		ctx:    ctx,
		cancel: cancel,
		ps:     ps,
		logger: logger,
	}, nil
}

func (m *Manager) ListenForCommands() {
	slog.Info("starting command listener")
	m.logger.LogServiceLifecycle("listening_for_commands", map[string]interface{}{
		"queue_name": "terminal",
	})
	queue.Listen("terminal", m.handleCommand)
}

func (m *Manager) handleCommand(msg string) {
	var cmd Command
	if err := json.Unmarshal([]byte(msg), &cmd); err != nil {
		slog.Error("failed to parse command")
		m.logger.LogTerminalError("PARSE_ERROR", "Failed to parse command message")
		return
	}

	// Skip empty commands (used for connection test)
	if cmd.Command == "" {
		if cmd.ResponseQueue != "" {
			if err := queue.Send(cmd.ResponseQueue, "Connected"); err != nil {
				slog.Warn("failed to send connection confirmation")
				m.logger.LogQueueOperation("send_connection_confirmation", false, nil)
			} else {
				m.logger.LogQueueOperation("send_connection_confirmation", true, nil)
			}
		}
		return
	}

	m.logger.LogTerminalEvent("command_received", "Command received", map[string]interface{}{
		"user_id": cmd.UserID,
	})

	response := m.executeCommand(cmd)

	if cmd.ResponseQueue != "" {
		if err := queue.Send(cmd.ResponseQueue, response); err != nil {
			slog.Warn("failed to send response")
			m.logger.LogQueueOperation("send_response", false, map[string]interface{}{
				"user_id": cmd.UserID,
			})
		} else {
			m.logger.LogQueueOperation("send_response", true, map[string]interface{}{
				"user_id":         cmd.UserID,
				"response_length": len(response),
			})
		}
	}
}

func (m *Manager) executeCommand(cmd Command) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	startTime := time.Now()

	if m.ps == nil {
		errorMsg := "Error: PowerShell session not initialized"
		m.logger.LogTerminalError("POWERSHELL_NOT_INITIALIZED", "PowerShell session not initialized")
		return errorMsg
	}

	// Format table output commands
	command := cmd.Command
	if strings.HasPrefix(command, "Get-") || strings.Contains(command, "| Format-Table") {
		if !strings.Contains(command, "| Format-Table") && !strings.Contains(command, "|ft") {
			command += " | Format-Table -AutoSize | Out-String -Width 120"
		}
	}

	output, err := m.ps.Execute(command)
	duration := time.Since(startTime)

	if err != nil {
		slog.Warn("failed to execute command")
		response := fmt.Sprintf("Error executing command: %v: %s", err, output)
		m.logger.LogCommandExecution(cmd.UserID, duration, false, len(output), command != cmd.Command)
		return response
	}

	m.logger.LogCommandExecution(cmd.UserID, duration, true, len(output), command != cmd.Command)

	return output
}

func (m *Manager) Shutdown() {
	slog.Info("shutting down terminal manager")
	m.logger.LogServiceLifecycle("shutting_down", map[string]interface{}{
		"service": "sirius-terminal",
	})
	m.cancel()
}
