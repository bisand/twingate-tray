package app

const (
	// AppName is the application name
	AppName = "Twingate Tray"

	// Version is the current application version
	Version = "1.0.0"

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
	return AppName + " v" + Version
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
