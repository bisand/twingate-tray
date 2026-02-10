package main

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"
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
	Interface      string
	IPAddress      string
	DNSServers     string
	DNSDomain      string
	Routes         string
	Resources      string
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
		Interface:      "-",
		IPAddress:      "-",
		DNSServers:     "-",
		DNSDomain:      "-",
		Routes:         "-",
		Resources:      "-",
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
		// Skip header line, parse data lines
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
		// First line is like "twingate 2025.342.178568 | 0.178.0"
		firstLine := strings.SplitN(strings.TrimSpace(out), "\n", 2)[0]
		if strings.HasPrefix(strings.ToLower(firstLine), "twingate") {
			info.ClientVersion = strings.TrimSpace(firstLine)
		}
	}

	// 4. Network interface info (sdwan0 is the Twingate interface)
	if out, err := runCommandOutput("ip", "addr", "show", "sdwan0"); err == nil {
		info.Interface = "sdwan0"
		// Extract inet address
		re := regexp.MustCompile(`inet\s+(\S+)`)
		if m := re.FindStringSubmatch(out); len(m) > 1 {
			info.IPAddress = m[1]
		}
	}

	// 5. DNS info
	if out, err := runCommandOutput("resolvectl", "status", "sdwan0"); err == nil {
		lines := strings.Split(out, "\n")
		var dnsServers []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DNS Servers:") || strings.HasPrefix(line, "Current DNS Server:") {
				// May have multiple servers on same line or continuation lines
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					servers := strings.TrimSpace(parts[1])
					if servers != "" && !contains(dnsServers, servers) {
						dnsServers = append(dnsServers, servers)
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
			// Extract just the destination network/host
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
		var resources []string
		// Skip header line
		for _, line := range lines[1:] {
			fields := splitTSV(line)
			if len(fields) >= 2 {
				name := strings.TrimSpace(fields[0])
				addr := strings.TrimSpace(fields[1])
				if name != "" {
					resources = append(resources, fmt.Sprintf("%s (%s)", name, addr))
				}
			}
		}
		if len(resources) > 0 {
			info.Resources = strings.Join(resources, "\n")
		}
	}

	// 8. Connected since (from systemd service start time)
	if out, err := runCommandOutput("systemctl", "show", "twingate", "--property=ActiveEnterTimestamp"); err == nil {
		// Format: ActiveEnterTimestamp=Mon 2026-02-10 18:52:02 CET
		parts := strings.SplitN(strings.TrimSpace(out), "=", 2)
		if len(parts) == 2 && parts[1] != "" {
			if t, err := parseSystemdTimestamp(parts[1]); err == nil {
				duration := time.Since(t)
				info.ConnectedSince = fmt.Sprintf("%s (%s ago)", t.Format("2006-01-02 15:04:05"), formatDuration(duration))
			} else {
				info.ConnectedSince = parts[1]
			}
		}
	}

	return info
}

// showStatusDialog displays the connection information dialog using zenity
func showStatusDialog(info ConnectionInfo) {
	// Build the info text as a formatted table
	lines := []string{
		fmt.Sprintf("<b>Status:</b>\t%s", escapeMarkup(info.Status)),
		fmt.Sprintf("<b>Connected since:</b>\t%s", escapeMarkup(info.ConnectedSince)),
		"",
		fmt.Sprintf("<b>User:</b>\t%s", escapeMarkup(info.UserEmail)),
		fmt.Sprintf("<b>Network:</b>\t%s", escapeMarkup(info.Network)),
		fmt.Sprintf("<b>Network URL:</b>\t%s", escapeMarkup(info.NetworkURL)),
		"",
		fmt.Sprintf("<b>Interface:</b>\t%s", escapeMarkup(info.Interface)),
		fmt.Sprintf("<b>IP address:</b>\t%s", escapeMarkup(info.IPAddress)),
		fmt.Sprintf("<b>DNS servers:</b>\t%s", escapeMarkup(info.DNSServers)),
		fmt.Sprintf("<b>DNS domain:</b>\t%s", escapeMarkup(info.DNSDomain)),
		fmt.Sprintf("<b>Secure DNS:</b>\t%s", escapeMarkup(info.SecureDNS)),
		"",
		fmt.Sprintf("<b>Routes:</b>\t%s", escapeMarkup(info.Routes)),
		"",
		fmt.Sprintf("<b>Resources:</b>\t%s", escapeMarkup(info.Resources)),
		"",
		fmt.Sprintf("<b>Client version:</b>\t%s", escapeMarkup(info.ClientVersion)),
	}

	text := strings.Join(lines, "\n")

	cmd := exec.Command("zenity", "--info",
		"--title=Twingate Connection Information",
		"--text="+text,
		"--width=500",
		"--no-wrap",
	)
	if err := cmd.Run(); err != nil {
		log.Printf("Status dialog error: %v", err)
	}
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

// escapeMarkup escapes characters that are special in Pango markup
func escapeMarkup(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// parseSystemdTimestamp attempts to parse a systemd timestamp string
func parseSystemdTimestamp(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Try various formats systemd might use
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
