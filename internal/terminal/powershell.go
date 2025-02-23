package terminal

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type PowerShell struct {
	cmd *exec.Cmd
}

func NewPowerShell() (*PowerShell, error) {
	// Test PowerShell availability
	testCmd := exec.Command("pwsh", "-NoProfile", "-Command", "Get-Location")
	if err := testCmd.Run(); err != nil {
		return nil, fmt.Errorf("PowerShell not available: %v", err)
	}

	return &PowerShell{}, nil
}

func (p *PowerShell) Execute(command string) (string, error) {
	// Create PowerShell process with formatting
	formattedCmd := fmt.Sprintf(`
		$Host.UI.RawUI.BufferSize = New-Object Management.Automation.Host.Size(120, 50)
		$Host.UI.RawUI.WindowSize = New-Object Management.Automation.Host.Size(120, 50)
		$OutputEncoding = [Console]::OutputEncoding = [Text.Encoding]::UTF8
		%s | Out-String -Width 120
	`, command)

	args := []string{
		"-NoProfile",
		"-NonInteractive",
		"-Command",
		formattedCmd,
	}

	cmd := exec.Command("pwsh", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), fmt.Errorf("command failed: %v: %s", err, stderr.String())
	}

	// Process the output
	output := stdout.String()
	output = strings.ReplaceAll(output, "\r\n", "\n") // Normalize line endings
	output = strings.TrimSpace(output)                // Remove leading/trailing whitespace

	// Ensure consistent line endings for terminal
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}

	return output, nil
}

func (p *PowerShell) Close() error {
	// Nothing to clean up in this implementation
	return nil
}
