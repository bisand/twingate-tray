package twingate

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.design/x/clipboard"
)

// ConnectionInfo holds all gathered connection details for the status dialog
type ConnectionInfo struct {
	Status         string
	ConnectedSince string
	Network        string
	NetworkURL     string
	UserEmail      string
	SecureDNS      string
	ClientVersion  string
	Hostname       string
	Interface      string
	InterfaceState string
	IPAddress      string
	IPv6Address    string
	MTU            string
	DNSServers     string
	DNSDomain      string
	Routes         string
	Resources      []ResourceEntry
	DaemonPID      string
	DaemonMemory   string
}

// ResourceEntry holds a single Twingate resource
type ResourceEntry struct {
	Name       string
	Address    string
	AuthStatus string
}

// gatherConnectionInfo collects connection information from various sources.
// Each field is gathered independently so partial failures don't block the dialog.
func gatherConnectionInfo() ConnectionInfo {
	info := ConnectionInfo{
		Status:         "Unknown",
		ConnectedSince: "-",
		Network:        "-",
		NetworkURL:     "-",
		UserEmail:      "-",
		SecureDNS:      "-",
		ClientVersion:  "-",
		Hostname:       "-",
		Interface:      "-",
		InterfaceState: "-",
		IPAddress:      "-",
		IPv6Address:    "-",
		MTU:            "-",
		DNSServers:     "-",
		DNSDomain:      "-",
		Routes:         "-",
		DaemonPID:      "-",
		DaemonMemory:   "-",
	}

	// Hostname
	if h, err := os.Hostname(); err == nil {
		info.Hostname = h
	}

	// 1. Status (verbose)
	if out, err := runCommandOutput("twingate", "status", "-v", "-d"); err == nil {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Online") {
				info.Status = "Online"
			} else if strings.HasPrefix(line, "Offline") || strings.HasPrefix(line, "offline") {
				info.Status = "Offline"
			}
			if strings.HasPrefix(line, "Secure DNS:") {
				info.SecureDNS = strings.TrimSpace(strings.TrimPrefix(line, "Secure DNS:"))
			}
		}
	}

	// 2. Account info
	if out, err := runCommandOutput("twingate", "account", "list", "-d"); err == nil {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, line := range lines[1:] {
			fields := splitTSV(line)
			if len(fields) >= 3 {
				info.UserEmail = fields[0]
				info.Network = fields[1]
				info.NetworkURL = fields[2]
			}
		}
	}

	// 3. Version
	if out, err := runCommandOutput("twingate", "version"); err == nil {
		firstLine := strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]
		if strings.HasPrefix(strings.ToLower(firstLine), "twingate") {
			info.ClientVersion = strings.TrimSpace(firstLine)
		}
	}

	// 4. Network interface info (sdwan0 is the Twingate interface)
	if out, err := runCommandOutput("ip", "addr", "show", "sdwan0"); err == nil {
		info.Interface = "sdwan0"

		// Extract inet address
		reV4 := regexp.MustCompile(`inet\s+(\S+)`)
		if m := reV4.FindStringSubmatch(out); len(m) > 1 {
			info.IPAddress = m[1]
		}

		// Extract inet6 address (link-local)
		reV6 := regexp.MustCompile(`inet6\s+(\S+)`)
		if m := reV6.FindStringSubmatch(out); len(m) > 1 {
			info.IPv6Address = m[1]
		}

		// Extract state from first line
		reState := regexp.MustCompile(`state\s+(\S+)`)
		if m := reState.FindStringSubmatch(out); len(m) > 1 {
			info.InterfaceState = m[1]
		}

		// Extract MTU
		reMTU := regexp.MustCompile(`mtu\s+(\d+)`)
		if m := reMTU.FindStringSubmatch(out); len(m) > 1 {
			info.MTU = m[1]
		}
	}

	// 5. DNS info
	if out, err := runCommandOutput("resolvectl", "status", "sdwan0"); err == nil {
		lines := strings.Split(out, "\n")
		var dnsServers []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DNS Servers:") || strings.HasPrefix(line, "Current DNS Server:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					servers := strings.TrimSpace(parts[1])
					for _, s := range strings.Fields(servers) {
						if !contains(dnsServers, s) {
							dnsServers = append(dnsServers, s)
						}
					}
				}
			}
			if strings.HasPrefix(line, "DNS Domain:") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					info.DNSDomain = strings.TrimSpace(parts[1])
				}
			}
		}
		if len(dnsServers) > 0 {
			info.DNSServers = strings.Join(dnsServers, ", ")
		}
	}

	// 6. Routes
	if out, err := runCommandOutput("ip", "route", "show", "dev", "sdwan0"); err == nil {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		var routes []string
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				routes = append(routes, fields[0])
			}
		}
		if len(routes) > 0 {
			info.Routes = strings.Join(routes, ", ")
		}
	}

	// 7. Resources
	if out, err := runCommandOutput("twingate", "resources", "-d"); err == nil {
		lines := strings.Split(strings.TrimSpace(out), "\n")
		for _, line := range lines[1:] {
			fields := splitTSV(line)
			if len(fields) >= 2 {
				entry := ResourceEntry{
					Name:    strings.TrimSpace(fields[0]),
					Address: strings.TrimSpace(fields[1]),
				}
				if len(fields) >= 4 {
					entry.AuthStatus = strings.TrimSpace(fields[3])
				} else if len(fields) >= 3 {
					entry.AuthStatus = strings.TrimSpace(fields[2])
				}
				if entry.Name != "" {
					info.Resources = append(info.Resources, entry)
				}
			}
		}
	}

	// 8. Connected since + daemon info (from systemd)
	if out, err := runCommandOutput("systemctl", "show", "twingate",
		"--property=ActiveEnterTimestamp,MainPID,MemoryCurrent"); err == nil {
		for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 || parts[1] == "" {
				continue
			}
			switch parts[0] {
			case "ActiveEnterTimestamp":
				if t, err := parseSystemdTimestamp(parts[1]); err == nil {
					duration := time.Since(t)
					info.ConnectedSince = fmt.Sprintf("%s (%s)",
						t.Format("2006-01-02 15:04:05"), formatDuration(duration))
				} else {
					info.ConnectedSince = parts[1]
				}
			case "MainPID":
				if parts[1] != "0" {
					info.DaemonPID = parts[1]
				}
			case "MemoryCurrent":
				if parts[1] != "[not set]" && parts[1] != "" {
					if mem, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64); err == nil {
						info.DaemonMemory = formatBytes(mem)
					}
				}
			}
		}
	}

	return info
}

// formatPlainText formats the ConnectionInfo as a plain-text string suitable for copying.
func (info *ConnectionInfo) formatPlainText() string {
	var b strings.Builder

	b.WriteString("=== Twingate Connection Information ===\n\n")

	b.WriteString(fmt.Sprintf("  Status:           %s\n", info.Status))
	b.WriteString(fmt.Sprintf("  Connected since:  %s\n", info.ConnectedSince))
	b.WriteString(fmt.Sprintf("  Hostname:         %s\n", info.Hostname))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  User:             %s\n", info.UserEmail))
	b.WriteString(fmt.Sprintf("  Network:          %s\n", info.Network))
	b.WriteString(fmt.Sprintf("  Network URL:      %s\n", info.NetworkURL))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  Interface:        %s (%s)\n", info.Interface, info.InterfaceState))
	b.WriteString(fmt.Sprintf("  IP address:       %s\n", info.IPAddress))
	b.WriteString(fmt.Sprintf("  IPv6 address:     %s\n", info.IPv6Address))
	b.WriteString(fmt.Sprintf("  MTU:              %s\n", info.MTU))
	b.WriteString(fmt.Sprintf("  DNS servers:      %s\n", info.DNSServers))
	b.WriteString(fmt.Sprintf("  DNS domain:       %s\n", info.DNSDomain))
	b.WriteString(fmt.Sprintf("  Secure DNS:       %s\n", info.SecureDNS))
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  Routes:           %s\n", info.Routes))
	b.WriteString("\n")

	b.WriteString("  Resources:\n")
	if len(info.Resources) == 0 {
		b.WriteString("    (none)\n")
	} else {
		for _, r := range info.Resources {
			if r.AuthStatus != "" && r.AuthStatus != "-" {
				b.WriteString(fmt.Sprintf("    %s  %s  [%s]\n", r.Name, r.Address, r.AuthStatus))
			} else {
				b.WriteString(fmt.Sprintf("    %s  %s\n", r.Name, r.Address))
			}
		}
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  Daemon PID:       %s\n", info.DaemonPID))
	b.WriteString(fmt.Sprintf("  Daemon memory:    %s\n", info.DaemonMemory))
	b.WriteString(fmt.Sprintf("  Client version:   %s\n", info.ClientVersion))

	return b.String()
}

// showStatusDialog displays the connection information dialog using zenity.
// Uses --text-info for a scrollable, selectable text view with a Copy button.
func showStatusDialog(info ConnectionInfo) {
	text := info.formatPlainText()

	for {
		cmd := exec.Command("zenity", "--text-info",
			"--title=Twingate Connection Information",
			"--width=550",
			"--height=500",
			"--font=monospace 10",
			"--ok-label=OK",
			"--extra-button=Copy to Clipboard",
		)
		cmd.Stdin = strings.NewReader(text)
		output, err := cmd.Output()

		// Check if the "Copy to Clipboard" button was clicked
		// zenity returns the extra button label on stdout with exit code 1
		buttonClicked := strings.TrimSpace(string(output))
		if buttonClicked == "Copy to Clipboard" {
			copyToClipboard(text)
			// Send notification
			cmd := exec.Command("notify-send", "-a", "Twingate Tray", "-t", "5000", "Twingate", "Connection info copied to clipboard")
			_ = cmd.Run() // Ignore errors - notification is optional
			// Re-show the dialog so the user can dismiss with OK
			continue
		}

		if err != nil && buttonClicked == "" {
			// Normal exit (OK or window close) or error
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 1 && exitErr.ExitCode() != 5 {
					log.Printf("Status dialog error: %v", err)
				}
			}
		}
		break
	}
}

// copyToClipboard copies text to the system clipboard using golang.design/x/clipboard.
func copyToClipboard(text string) {
	err := clipboard.Init()
	if err != nil {
		log.Printf("Failed to initialize clipboard: %v", err)
		return
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
}

// runCommandOutput executes a command and returns its combined output.
// This is a convenience wrapper around runCommand from twingate.go.
var runCommandOutput = runCommand

// splitTSV splits a tab-separated line into fields, trimming whitespace
func splitTSV(line string) []string {
	parts := strings.Split(line, "\t")
	var fields []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			fields = append(fields, trimmed)
		}
	}
	return fields
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

// parseSystemdTimestamp attempts to parse a systemd timestamp string
func parseSystemdTimestamp(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		"Mon 2006-01-02 15:04:05 MST",
		"Mon 2006-01-02 15:04:05 -0700",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse timestamp: %s", s)
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	hours := int(d.Hours())
	if hours < 24 {
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh %dm", hours, mins)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := hours / 24
	remainHours := hours % 24
	if remainHours > 0 {
		return fmt.Sprintf("%dd %dh", days, remainHours)
	}
	return fmt.Sprintf("%d days", days)
}

// formatBytes formats a byte count into a human-readable string
func formatBytes(bytes uint64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ShowConnectionInfo gathers and displays the connection information dialog
func ShowConnectionInfo() {
	info := gatherConnectionInfo()
	showStatusDialog(info)
}
