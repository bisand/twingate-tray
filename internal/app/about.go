package app

import (
	"os"
	"os/exec"
	"path/filepath"
)

// ShowAbout displays the About dialog using zenity with Pango markup
func ShowAbout() error {
	aboutText := GetAboutText()

	// Try to find the icon
	iconPath := findIconPath()

	args := []string{
		"--info",
		"--title=About",
		"--text=" + aboutText,
		"--width=450",
		"--height=300",
	}

	// Add icon if we found one
	if iconPath != "" {
		args = append(args, "--window-icon="+iconPath)
	}

	cmd := exec.Command("zenity", args...)
	return cmd.Run()
}

// findIconPath tries to locate the icon in various locations
func findIconPath() string {
	// Try system icon name first (works if icon is installed)
	if iconExists("twingate-tray") {
		return "" // Will use --icon-name instead
	}

	// Try relative paths from executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)

		// Try ../assets/twingate-tray.svg (development structure)
		devIcon := filepath.Join(execDir, "..", "assets", "twingate-tray.svg")
		if _, err := os.Stat(devIcon); err == nil {
			return devIcon
		}

		// Try ./assets/twingate-tray.svg
		localIcon := filepath.Join(execDir, "assets", "twingate-tray.svg")
		if _, err := os.Stat(localIcon); err == nil {
			return localIcon
		}
	}

	// Try current directory
	if _, err := os.Stat("assets/twingate-tray.svg"); err == nil {
		return "assets/twingate-tray.svg"
	}

	return ""
}

// iconExists checks if a system icon exists
func iconExists(iconName string) bool {
	// For now, always use file path if available
	// In the future, could check icon cache
	return false
}
