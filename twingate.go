package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// checkTwingateStatus returns true if connected to Twingate, false otherwise
func checkTwingateStatus() (bool, error) {
	output, err := runCommand("twingate", "status")
	if err != nil {
		return false, fmt.Errorf("twingate status command failed: %w", err)
	}

	// Check if output starts with "online"
	return strings.HasPrefix(strings.TrimSpace(output), "online"), nil
}

// handleConnect connects to Twingate
func handleConnect() error {
	// Try pkexec first, fall back to sudo
	if err := runPrivilegedCommand("twingate", "start"); err != nil {
		return fmt.Errorf("failed to start twingate: %w", err)
	}

	// Also try desktop-restart
	runCommand("twingate", "desktop-restart")

	return nil
}

// handleDisconnect disconnects from Twingate
func handleDisconnect() error {
	// Try pkexec first, fall back to sudo
	if err := runPrivilegedCommand("twingate", "stop"); err != nil {
		return fmt.Errorf("failed to stop twingate: %w", err)
	}

	// Also try desktop-stop
	runCommand("twingate", "desktop-stop")

	return nil
}

// runCommand executes a simple command and returns its output
func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// runPrivilegedCommand runs a command with elevated privileges
// Tries pkexec first, then falls back to sudo
func runPrivilegedCommand(name string, args ...string) error {
	// Try pkexec first (preferred on most modern Linux DEs)
	cmd := exec.Command("pkexec", append([]string{name}, args...)...)
	if err := cmd.Run(); err == nil {
		return nil
	}

	// Fall back to sudo
	cmd = exec.Command("sudo", append([]string{name}, args...)...)
	return cmd.Run()
}
