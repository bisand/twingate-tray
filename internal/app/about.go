package app

import (
	"os"
	"os/exec"
	"path/filepath"
)

// ShowAbout displays the About dialog
func ShowAbout() error {
	aboutText := GetAboutText()

	// Try yad first (supports large body images with --image)
	if _, err := exec.LookPath("yad"); err == nil {
		return showAboutWithYad(aboutText)
	}

	// Fall back to zenity (smaller icon only)
	return showAboutWithZenity(aboutText)
}

// showAboutWithYad shows About dialog using yad (supports large images)
func showAboutWithYad(aboutText string) error {
	args := []string{
		"--info",
		"--title=About",
		"--text=" + aboutText,
		"--width=480",
		"--height=400",
		"--image-on-top",
		"--center",
		"--fixed",
		"--buttons-layout=spread",
		"--button=    OK    :0",
	}

	// Find icon for display (prefers 96x96 SVG)
	iconPath := findAboutIconPath()
	if iconPath != "" {
		args = append(args, "--image="+iconPath)
	}

	cmd := exec.Command("yad", args...)
	return cmd.Run()
}

// showAboutWithZenity shows About dialog using zenity (small icon only)
func showAboutWithZenity(aboutText string) error {
	args := []string{
		"--info",
		"--title=About",
		"--text=" + aboutText,
		"--width=500",
		"--height=350",
	}

	// Add window icon (titlebar only, zenity doesn't support large body images)
	if isIconInstalled() {
		args = append(args, "--window-icon=twingate-tray")
	} else {
		iconPath := findAboutIconPath()
		if iconPath != "" {
			args = append(args, "--window-icon="+iconPath)
		}
	}

	cmd := exec.Command("zenity", args...)
	return cmd.Run()
}

// isIconInstalled checks if the icon is installed in the system theme
func isIconInstalled() bool {
	// Check if icon exists in common icon theme locations
	iconPaths := []string{
		"/usr/share/icons/hicolor/scalable/apps/twingate-tray.svg",
		"/usr/share/icons/hicolor/128x128/apps/twingate-tray.png",
		"/usr/share/icons/hicolor/256x256/apps/twingate-tray.png",
	}

	for _, path := range iconPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// findAboutIconPath tries to locate the icon for About dialog
func findAboutIconPath() string {
	// Look for 96x96 SVG first (properly sized for About dialog)
	iconPaths := []string{
		"assets/twingate-tray-96.svg",
		"assets/twingate-tray.svg",
	}

	for _, path := range iconPaths {
		if _, err := os.Stat(path); err == nil {
			abs, _ := filepath.Abs(path)
			return abs
		}
	}

	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)

		// Try ../assets/twingate-tray-96.svg (development structure)
		devIcon96 := filepath.Join(execDir, "..", "assets", "twingate-tray-96.svg")
		if _, err := os.Stat(devIcon96); err == nil {
			abs, _ := filepath.Abs(devIcon96)
			return abs
		}

		// Try ../assets/twingate-tray.svg (development structure)
		devIcon := filepath.Join(execDir, "..", "assets", "twingate-tray.svg")
		if _, err := os.Stat(devIcon); err == nil {
			abs, _ := filepath.Abs(devIcon)
			return abs
		}

		// Try ./assets/twingate-tray-96.svg
		localIcon96 := filepath.Join(execDir, "assets", "twingate-tray-96.svg")
		if _, err := os.Stat(localIcon96); err == nil {
			abs, _ := filepath.Abs(localIcon96)
			return abs
		}

		// Try ./assets/twingate-tray.svg
		localIcon := filepath.Join(execDir, "assets", "twingate-tray.svg")
		if _, err := os.Stat(localIcon); err == nil {
			abs, _ := filepath.Abs(localIcon)
			return abs
		}
	}

	return ""
}
