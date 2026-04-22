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
		slog.Error("failed to parse command", "error", err)
		m.logger.LogTerminalError("", "", "PARSE_ERROR", "Failed to parse command message", err)
		return
	}

	// Skip empty commands (used for connection test)
	if cmd.Command == "" {
		if cmd.ResponseQueue != "" {
			if err := queue.Send(cmd.ResponseQueue, "Connected"); err != nil {
				slog.Warn("failed to send connection confirmation", "queue", cmd.ResponseQueue, "error", err)
				m.logger.LogQueueOperation("send_connection_confirmation", cmd.ResponseQueue, false, map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				m.logger.LogQueueOperation("send_connection_confirmation", cmd.ResponseQueue, true, map[string]interface{}{
					"message": "Connected",
				})
			}
		}
		return
	}

	// Log command received
	m.logger.LogTerminalEvent("command_received", fmt.Sprintf("Command received from user %s", cmd.UserID), map[string]interface{}{
		"user_id":        cmd.UserID,
		"command":        cmd.Command,
		"response_queue": cmd.ResponseQueue,
		"timestamp":      cmd.Timestamp,
	})

	response := m.executeCommand(cmd)

	if cmd.ResponseQueue != "" {
		if err := queue.Send(cmd.ResponseQueue, response); err != nil {
			slog.Warn("failed to send response", "queue", cmd.ResponseQueue, "error", err)
			m.logger.LogQueueOperation("send_response", cmd.ResponseQueue, false, map[string]interface{}{
				"user_id": cmd.UserID,
				"error":   err.Error(),
			})
		} else {
			m.logger.LogQueueOperation("send_response", cmd.ResponseQueue, true, map[string]interface{}{
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
		m.logger.LogTerminalError(cmd.UserID, cmd.Command, "POWERSHELL_NOT_INITIALIZED", "PowerShell session not initialized", fmt.Errorf("PowerShell session is nil"))
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
		slog.Warn("failed to execute command", "command", cmd.Command, "error", err)
		errorMsg := fmt.Sprintf("Error executing command: %v", err)
		m.logger.LogCommandExecution(cmd.UserID, cmd.Command, duration, false, len(errorMsg), map[string]interface{}{
			"error": err.Error(),
		})
		return errorMsg
	}

	// Log successful command execution
	m.logger.LogCommandExecution(cmd.UserID, cmd.Command, duration, true, len(output), map[string]interface{}{
		"formatted_command": command != cmd.Command,
		"output_length":     len(output),
	})

	return output
}

func (m *Manager) Shutdown() {
	slog.Info("shutting down terminal manager")
	m.logger.LogServiceLifecycle("shutting_down", map[string]interface{}{
		"service": "sirius-terminal",
	})
	m.cancel()
}
