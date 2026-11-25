package terminal

import (
	"fmt"
	"time"

	"github.com/SiriusScan/go-api/sirius/logging"
)

// LoggingClient provides a centralized way to send structured logs to the API
// This is now a wrapper around the SDK's global logging client.
type LoggingClient struct{}

// NewLoggingClient creates a new LoggingClient instance.
// It ensures the SDK's global client is initialized.
func NewLoggingClient() *LoggingClient {
	// The SDK's Init() function is idempotent, safe to call multiple times.
	// It will only initialize if not already initialized.
	logging.Init()
	return &LoggingClient{}
}

// LogTerminalEvent logs a general event related to terminal operations
func (lc *LoggingClient) LogTerminalEvent(eventType, message string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["event_type"] = eventType
	logging.Log("sirius-terminal", "terminal-manager", logging.LogLevelInfo, message, metadata, map[string]interface{}{"type": "business_event"})
}

// LogCommandExecution logs the execution of a terminal command
func (lc *LoggingClient) LogCommandExecution(userID string, duration time.Duration, success bool, outputLength int, formatted bool) {
	metadata := map[string]interface{}{
		"user_id":           userID,
		"duration_ms":       duration.Milliseconds(),
		"success":           success,
		"output_length":     outputLength,
		"formatted_command": formatted,
	}
	
	level := logging.LogLevelInfo
	if !success {
		level = logging.LogLevelError
	}
	
	message := fmt.Sprintf("Command execution %s", messageFromSuccess(success))
	logging.Log("sirius-terminal", "command-execution", level, message, metadata, map[string]interface{}{"type": "performance_metric"})
}

// LogTerminalError logs an error related to terminal operations
func (lc *LoggingClient) LogTerminalError(errorCode, message string) {
	metadata := map[string]interface{}{"error_code": errorCode}
	logging.Log("sirius-terminal", "terminal-manager", logging.LogLevelError, message, metadata, map[string]interface{}{"type": "error"})
}

// LogPowerShellInitialization logs PowerShell session initialization
func (lc *LoggingClient) LogPowerShellInitialization(success bool, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["success"] = success
	
	level := logging.LogLevelInfo
	if !success {
		level = logging.LogLevelError
	}
	
	message := fmt.Sprintf("PowerShell initialization %s", messageFromSuccess(success))
	logging.Log("sirius-terminal", "powershell-init", level, message, metadata, map[string]interface{}{"type": "system_event"})
}

// LogQueueOperation logs queue-related operations
func (lc *LoggingClient) LogQueueOperation(operation string, success bool, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["operation"] = operation
	metadata["success"] = success
	
	level := logging.LogLevelInfo
	if !success {
		level = logging.LogLevelError
	}
	
	message := fmt.Sprintf("Queue operation %s: %s", operation, messageFromSuccess(success))
	logging.Log("sirius-terminal", "queue-operation", level, message, metadata, map[string]interface{}{"type": "system_event"})
}

// LogServiceLifecycle logs service lifecycle events
func (lc *LoggingClient) LogServiceLifecycle(event string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["lifecycle_event"] = event
	
	logging.Log("sirius-terminal", "service-lifecycle", logging.LogLevelInfo, fmt.Sprintf("Service %s", event), metadata, map[string]interface{}{"type": "system_event"})
}

func messageFromSuccess(success bool) string {
	if success {
		return "completed successfully"
	}
	return "failed"
}
