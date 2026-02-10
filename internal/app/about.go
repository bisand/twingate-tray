package app

import (
	"os/exec"
)

// ShowAbout displays the About dialog using zenity with Pango markup
func ShowAbout() error {
	aboutText := GetAboutText()

	// Zenity supports Pango markup by default (unless --no-markup is used)
	cmd := exec.Command("zenity", "--info",
		"--title=About",
		"--text="+aboutText,
		"--icon-name=twingate-tray",
		"--width=450",
		"--height=300")

	return cmd.Run()
}
