package twingate

import (
	"fmt"
	"os/exec"
	"strings"
)

// NetworkInfo represents basic network information
type NetworkInfo struct {
	Name string
	URL  string
}

// CheckStatus returns true if connected to Twingate, false otherwise
func CheckStatus() (bool, error) {
	output, err := runCommand("twingate", "status")
	if err != nil {
		return false, fmt.Errorf("twingate status command failed: %w", err)
	}

	// Check if output starts with "online"
	return strings.HasPrefix(strings.TrimSpace(output), "online"), nil
}

// GetNetworkInfo retrieves the current network name and URL
func GetNetworkInfo() (*NetworkInfo, error) {
	output, err := runCommand("twingate", "account", "list", "-d")
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return &NetworkInfo{Name: "-", URL: "-"}, nil
	}

	// Parse first data line (skip header)
	fields := strings.Split(lines[1], "\t")
	info := &NetworkInfo{Name: "-", URL: "-"}

	if len(fields) >= 2 {
		info.Name = strings.TrimSpace(fields[1])
	}
	if len(fields) >= 3 {
		info.URL = strings.TrimSpace(fields[2])
	}

	return info, nil
}

// IsAutoConnectEnabled checks if the Twingate service is set to start automatically
func IsAutoConnectEnabled() bool {
	output, err := runCommand("systemctl", "is-enabled", "twingate")
	if err != nil {
		return false
	}
	return strings.TrimSpace(output) == "enabled"
}

// SetAutoConnect enables or disables auto-connect by enabling/disabling the systemd service
func SetAutoConnect(enabled bool) error {
	var cmd *exec.Cmd
	if enabled {
		cmd = exec.Command("pkexec", "systemctl", "enable", "twingate")
	} else {
		cmd = exec.Command("pkexec", "systemctl", "disable", "twingate")
	}

	if err := cmd.Run(); err != nil {
		// Try with sudo as fallback
		if enabled {
			cmd = exec.Command("sudo", "systemctl", "enable", "twingate")
		} else {
			cmd = exec.Command("sudo", "systemctl", "disable", "twingate")
		}
		return cmd.Run()
	}

	return nil
}

// Connect connects to Twingate
func Connect() error {
	// Try pkexec first, fall back to sudo
	if err := runPrivilegedCommand("twingate", "start"); err != nil {
		return fmt.Errorf("failed to start twingate: %w", err)
	}

	// Also try desktop-restart
	runCommand("twingate", "desktop-restart")

	return nil
}

// Disconnect disconnects from Twingate
func Disconnect() error {
	// Try pkexec first, fall back to sudo
	if err := runPrivilegedCommand("twingate", "stop"); err != nil {
		return fmt.Errorf("failed to stop twingate: %w", err)
	}

	// Also try desktop-stop
	runCommand("twingate", "desktop-stop")

	return nil
}

// GenerateDiagnosticReport generates a diagnostic report
func GenerateDiagnosticReport() error {
	cmd := exec.Command("twingate", "report")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate diagnostic report: %w\nOutput: %s", err, string(output))
	}

	// Show success notification with report location
	// The report command typically outputs the file path
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
