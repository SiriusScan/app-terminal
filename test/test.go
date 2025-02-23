package main

import (
	"fmt"
	"log"
	"os/exec"
)

func executePowerShell(command string) (string, error) {
	// Basic PowerShell execution
	cmd := exec.Command("pwsh", "-Command", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to execute command: %v", err)
	}
	return string(output), nil
}

func main() {
	log.Println("Starting PowerShell test...")

	// Test commands
	commands := []string{
		"Get-Location",
		"Get-Date",
		"Get-Process | Select-Object -First 5",
		"$PSVersionTable",
	}

	for _, cmd := range commands {
		log.Printf("Executing: %s", cmd)
		output, err := executePowerShell(cmd)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		log.Printf("Output:\n%s", output)
		log.Println("---")
	}
}
