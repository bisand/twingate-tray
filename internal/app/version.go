package app

// Version information
// These variables can be overridden at build time using -ldflags
var (
	// Version is the current application version
	// Override with: go build -ldflags "-X github.com/bisand/twingate-tray/internal/app.Version=1.2.3"
	Version = "dev"

	// GitCommit is the git commit SHA
	GitCommit = "unknown"

	// BuildDate is when the binary was built
	BuildDate = "unknown"
)

const (
	// AppName is the application name
	AppName = "Twingate Tray"

	// Description is a short description of the application
	Description = "System Tray Indicator for Twingate VPN on Linux"

	// License is the software license
	License = "MIT License"

	// Author is the primary author/maintainer
	Author = "Community Project"

	// Repository is the source code repository URL
	Repository = "https://github.com/bisand/twingate-tray"

	// Credits lists contributors and acknowledgments
	Credits = `Built with Go and D-Bus
StatusNotifierItem protocol
Integrates with Twingate CLI
System dialogs via zenity
Desktop notifications`
)

// GetFullVersion returns the formatted version string
func GetFullVersion() string {
	version := Version
	// Remove leading 'v' if present since we'll add it
	if len(version) > 0 && version[0] == 'v' {
		version = version[1:]
	}
	return AppName + " v" + version
}

// GetVersionInfo returns detailed version information
func GetVersionInfo() string {
	info := "Version: " + Version
	if GitCommit != "unknown" {
		info += "\nCommit: " + GitCommit
	}
	if BuildDate != "unknown" {
		info += "\nBuilt: " + BuildDate
	}
	return info
}

// GetAboutText returns the complete about text for display
func GetAboutText() string {
	return "<b><big>" + AppName + "</big></b>\n\n" +
		"<b>Version:</b> " + Version + "\n\n" +
		Description + "\n\n" +
		"<b>License:</b> " + License + "\n" +
		Author + "\n\n" +
		"<b>Repository:</b>\n" +
		"<a href=\"" + Repository + "\">" + Repository + "</a>\n\n" +
		"<small>This program comes with absolutely no warranty.</small>"
}

// GetCreditsText returns credits information
func GetCreditsText() string {
	return "<b>Credits</b>\n\n" + Credits
}
