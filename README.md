# Sirius Terminal Application

The Sirius Terminal is a web-based PowerShell terminal that provides secure remote command execution through a message queue system.

(This README generated by AI)

## Architecture

### Frontend (Sirius/sirius-ui)

- Built with Next.js and TypeScript
- Uses xterm.js for terminal emulation
- Implements the Catppuccin Mocha color scheme
- Handles command history and keyboard shortcuts
- Communicates with backend via tRPC and RabbitMQ

Key files:

- `src/pages/terminal.tsx`: Main terminal interface
- `src/server/api/routers/terminal.ts`: tRPC endpoint handling
- `src/server/api/routers/queue.ts`: RabbitMQ message handling

### Backend (minor-projects/app-terminal)

- Written in Go
- Executes PowerShell commands securely
- Uses RabbitMQ for message queuing
- Implements mutex-protected command execution

Key files:

- `cmd/main.go`: Application entry point
- `internal/terminal/manager.go`: Command handling and queue management
- `internal/terminal/powershell.go`: PowerShell execution interface

## Features

### Terminal Interface

- Command history navigation (Up/Down arrows)
- Line editing (Backspace)
- Keyboard shortcuts:
  - Ctrl+C: Cancel current command
  - Ctrl+L: Clear screen
- Command completion
- Error handling with color-coded output

### Built-in Commands

```
typescript
const LOCAL_COMMANDS = {
help: // Show available commands
clear: // Clear terminal screen
history: // Show command history
exit: // Close session
version: // Show terminal version
}
```

### PowerShell Integration

- Secure command execution
- Output formatting
- Error handling
- Session management

## Message Flow

1. User enters command in terminal
2. Frontend sends command via tRPC
3. Command is queued in RabbitMQ
4. Backend processes command:
   - Validates command
   - Executes in PowerShell
   - Formats output
   - Returns response via queue
5. Frontend displays formatted response

## Configuration

### Frontend Terminal Settings

```typescript
{
  cursorBlink: true,
  fontSize: 14,
  fontFamily: "JetBrains Mono, Menlo, Monaco, 'Courier New', monospace",
  theme: MOCHA_THEME,
  letterSpacing: 0,
  lineHeight: 1.2,
  scrollback: 5000,
  cursorStyle: "block",
  cursorWidth: 2,
  rows: 40,
  cols: 100,
}
```

### PowerShell Configuration

```powershell
$Host.UI.RawUI.BufferSize = New-Object Management.Automation.Host.Size(120, 50)
$Host.UI.RawUI.WindowSize = New-Object Management.Automation.Host.Size(120, 50)
$OutputEncoding = [Console]::OutputEncoding = [Text.Encoding]::UTF8
```

## Security Considerations

1. Command Validation

   - All commands are executed in isolated PowerShell sessions
   - Session cleanup after execution

2. Queue Security

   - Non-persistent messages
   - Queue purging between commands
   - Timeout handling

3. Error Handling
   - Command execution errors
   - Queue connection errors
   - PowerShell session errors

## Development

### Prerequisites

- Node.js 16+
- Go 1.19+
- PowerShell Core (pwsh)
- RabbitMQ

### Local Setup

1. Start RabbitMQ
2. Run frontend development server
3. Start terminal backend service

### Testing

- Frontend component tests
- Backend integration tests
- Queue communication tests

## Deployment

The application is designed to run in a containerized environment with:

- Frontend container
- Backend container
- RabbitMQ container
- Shared network for communication

## Error Handling

The application implements comprehensive error handling:

1. PowerShell execution errors
2. Queue communication errors
3. Command parsing errors
4. Session management errors

Each error type has specific handling and user feedback mechanisms.
