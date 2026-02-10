package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bisand/twingate-tray/internal/app"
	"github.com/bisand/twingate-tray/internal/tray"
	"github.com/bisand/twingate-tray/internal/twingate"
)

var (
	appState   *app.AppState
	systemTray *tray.SystemTray
)

func main() {
	if len(os.Args) > 1 {
		// CLI mode - handle commands
		handleCLI(os.Args[1:])
		return
	}

	// Daemon mode with system tray
	startDaemon()
}

func startDaemon() {
	log.Println("Starting Twingate tray...")

	appState = app.NewAppState()

	// Detect auto-connect status from systemd
	autoConnectEnabled := twingate.IsAutoConnectEnabled()
	log.Printf("Auto-connect detected: %v", autoConnectEnabled)

	// Initialize system tray with all callback handlers
	var err error
	systemTray, err = tray.NewSystemTray(tray.CallbackHandlers{
		OnConnect:          handleConnect,
		OnDisconnect:       handleDisconnect,
		OnConnectionInfo:   handleConnectionInfo,
		OnRefreshStatus:    handleRefreshStatus,
		OnExitNodeStart:    handleExitNodeStart,
		OnExitNodeStop:     handleExitNodeStop,
		OnExitNodeList:     handleExitNodeList,
		OnExitNodeSwitch:   handleExitNodeSwitch,
		OnResourcesShow:    handleResourcesShow,
		OnOpenWebAdmin:     handleOpenWebAdmin,
		OnDiagReport:       handleDiagnosticReport,
		OnAutoConnToggle:   handleAutoConnectToggle,
		OnMenuOpening:      handleMenuOpening,
		OnQuit:             handleQuit,
		InitialAutoConnect: autoConnectEnabled,
	})

	if err != nil {
		log.Printf("Error: Could not initialize system tray: %v", err)
		os.Exit(1)
	}

	err = systemTray.Start()
	if err != nil {
		log.Printf("Error: Could not start system tray: %v", err)
		os.Exit(1)
	}

	log.Println("System tray initialized")

	// Setup signal handling for clean shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v - shutting down...", sig)
		cleanup()
		os.Exit(0)
	}()

	// Start status monitor in background
	go monitorStatus()

	// Start connection timer updater
	go updateConnectionTimer()

	// Keep running
	select {}
}

func cleanup() {
	log.Println("Cleaning up...")
	if systemTray != nil {
		systemTray.Stop()
	}
	log.Println("Cleanup complete")
}

func handleCLI(args []string) {
	if len(args) == 0 {
		printUsage()
		return
	}

	switch args[0] {
	case "status":
		connected, err := twingate.CheckStatus()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if connected {
			fmt.Println("connected")
		} else {
			fmt.Println("disconnected")
		}

	case "connect":
		if err := twingate.Connect(); err != nil {
			fmt.Printf("Error connecting: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Connection initiated")

	case "disconnect":
		if err := twingate.Disconnect(); err != nil {
			fmt.Printf("Error disconnecting: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Disconnection initiated")

	case "daemon":
		// Start as daemon with system tray
		startDaemon()

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Twingate Tray - System tray indicator for Twingate

Usage:
  twingate-tray                    # Run with system tray
  twingate-tray status             # Check connection status
  twingate-tray connect            # Connect to Twingate
  twingate-tray disconnect         # Disconnect from Twingate
  twingate-tray daemon             # Start as daemon with system tray
  twingate-tray help               # Show this help message`)
}

func monitorStatus() {
	ticker := time.NewTicker(app.StatusPollInterval)
	defer ticker.Stop()

	var prevConnected *bool

	for range ticker.C {
		updateStatus()

		// Log status changes
		connected := appState.IsConnected()
		if prevConnected == nil || *prevConnected != connected {
			if connected {
				log.Println("Status: Connected to Twingate")
				sendNotification("Twingate Connected", "You are now connected to Twingate")

				// Fetch and update network info on connect
				updateNetworkInfo()
			} else {
				log.Println("Status: Disconnected from Twingate")
				sendNotification("Twingate Disconnected", "You are now disconnected from Twingate")
			}
			prevConnected = &connected
		}
	}
}

// updateStatus checks current Twingate status and updates the app state
func updateStatus() {
	connected, err := twingate.CheckStatus()

	appState.SetConnected(connected)

	if err != nil {
		appState.SetLastError(err.Error())
		log.Printf("Status check error: %v", err)
	} else {
		appState.SetLastError("")
	}

	// Update system tray icon
	if systemTray != nil {
		systemTray.UpdateStatus(connected)
	}
}

// updateNetworkInfo fetches and updates network information in the tray
func updateNetworkInfo() {
	info, err := twingate.GetNetworkInfo()
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
		return
	}

	appState.SetNetworkName(info.Name)
	appState.SetNetworkURL(info.URL)

	if systemTray != nil {
		systemTray.UpdateNetworkInfo(info.Name, info.URL)
	}
}

// updateConnectionTimer updates the connection time display in the menu
func updateConnectionTimer() {
	ticker := time.NewTicker(30 * time.Second) // Update every 30 seconds
	defer ticker.Stop()

	for range ticker.C {
		if !appState.IsConnected() {
			continue
		}

		duration := appState.GetConnectionDuration()
		if duration == 0 {
			continue
		}

		// Format duration nicely
		timeStr := formatDuration(duration)

		if systemTray != nil {
			systemTray.UpdateConnectionTime(timeStr)
		}
	}
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	hours := int(d.Hours())
	if hours < 24 {
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh %dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 24
	remainHours := hours % 24
	if remainHours > 0 {
		return fmt.Sprintf("%dd %dh", days, remainHours)
	}
	return fmt.Sprintf("%dd", days)
}

func sendNotification(title, body string) {
	// Try to send via notify-send if available
	cmd := exec.Command("notify-send", "-a", "Twingate Tray", "-t", fmt.Sprintf("%d", app.NotificationTimeout), title, body)
	// Ignore errors - notification is optional
	_ = cmd.Run()
}

// Handler functions for tray callbacks

func handleConnect() {
	if err := twingate.Connect(); err != nil {
		log.Printf("Connect failed: %v", err)
		sendNotification("Connection Failed", fmt.Sprintf("Failed to connect: %v", err))
	}
}

func handleDisconnect() {
	if err := twingate.Disconnect(); err != nil {
		log.Printf("Disconnect failed: %v", err)
		sendNotification("Disconnection Failed", fmt.Sprintf("Failed to disconnect: %v", err))
	}
}

func handleConnectionInfo() {
	twingate.ShowConnectionInfo()
}

func handleRefreshStatus() {
	log.Println("Refreshing status...")
	updateStatus()
	updateNetworkInfo()
	sendNotification("Status Refreshed", "Twingate status has been refreshed")
}

func handleExitNodeStart() {
	log.Println("Starting exit node...")
	if err := twingate.StartExitNode(); err != nil {
		log.Printf("Failed to start exit node: %v", err)
		sendNotification("Exit Node Failed", fmt.Sprintf("Failed to start exit node: %v", err))
	} else {
		sendNotification("Exit Node Started", "All traffic is now routed through Twingate")
	}
}

func handleExitNodeStop() {
	log.Println("Stopping exit node...")
	if err := twingate.StopExitNode(); err != nil {
		log.Printf("Failed to stop exit node: %v", err)
		sendNotification("Exit Node Failed", fmt.Sprintf("Failed to stop exit node: %v", err))
	} else {
		sendNotification("Exit Node Stopped", "Split tunnel mode restored")
	}
}

func handleExitNodeList() {
	log.Println("Showing exit node list...")
	status, err := twingate.GetExitNodeStatus()
	if err != nil {
		log.Printf("Failed to get exit node status: %v", err)
		sendNotification("Exit Node Error", fmt.Sprintf("Failed to get exit nodes: %v", err))
		return
	}

	if len(status.AvailableNodes) == 0 {
		exec.Command("zenity", "--info", "--title=Exit Nodes", "--text=No exit nodes available", "--width=300").Run()
		return
	}

	// Build menu options for zenity
	var options []string
	if status.Enabled {
		options = append(options, "Stop Exit Node")
		options = append(options, "---")
	} else {
		options = append(options, "Start Exit Node")
		options = append(options, "---")
	}

	for _, node := range status.AvailableNodes {
		label := node
		if status.Enabled && node == status.CurrentNode {
			label = node + " (active)"
		}
		options = append(options, label)
	}

	// Show zenity menu
	cmd := exec.Command("zenity", "--list", "--title=Exit Nodes", "--text=Select an action:",
		"--column=Option", "--width=400", "--height=300")
	cmd.Args = append(cmd.Args, options...)

	output, err := cmd.Output()
	if err != nil {
		// User cancelled or closed dialog
		return
	}

	selected := string(output)
	selected = selected[:len(selected)-1] // Remove trailing newline

	if selected == "Start Exit Node" {
		handleExitNodeStart()
	} else if selected == "Stop Exit Node" {
		handleExitNodeStop()
	} else {
		// Extract node name (remove " (active)" if present)
		nodeName := selected
		if idx := len(selected) - 9; idx > 0 && selected[idx:] == " (active)" {
			nodeName = selected[:idx]
		}
		// Switch to the selected node
		if err := twingate.SwitchExitNode(nodeName); err != nil {
			log.Printf("Failed to switch exit node: %v", err)
			sendNotification("Exit Node Failed", fmt.Sprintf("Failed to switch to %s: %v", nodeName, err))
		} else {
			sendNotification("Exit Node Switched", fmt.Sprintf("Now using exit node: %s", nodeName))
		}
	}
}

func handleExitNodeSwitch() {
	log.Println("Switching exit node...")
	status, err := twingate.GetExitNodeStatus()
	if err != nil {
		log.Printf("Failed to get exit node status: %v", err)
		return
	}

	if len(status.AvailableNodes) == 0 {
		sendNotification("Exit Node Error", "No exit nodes available")
		return
	}

	// Show zenity dialog to select node
	cmd := exec.Command("zenity", "--list", "--title=Switch Exit Node",
		"--text=Select an exit node:", "--column=Node", "--width=400", "--height=300")
	cmd.Args = append(cmd.Args, status.AvailableNodes...)

	output, err := cmd.Output()
	if err != nil {
		// User cancelled
		return
	}

	nodeName := string(output)
	nodeName = nodeName[:len(nodeName)-1] // Remove trailing newline

	if err := twingate.SwitchExitNode(nodeName); err != nil {
		log.Printf("Failed to switch exit node: %v", err)
		sendNotification("Exit Node Failed", fmt.Sprintf("Failed to switch to %s: %v", nodeName, err))
	} else {
		sendNotification("Exit Node Switched", fmt.Sprintf("Now using exit node: %s", nodeName))
	}
}

func handleResourcesShow() {
	log.Println("Showing resources...")
	resources, err := twingate.GetResources()
	if err != nil {
		log.Printf("Failed to get resources: %v", err)
		sendNotification("Resources Error", fmt.Sprintf("Failed to get resources: %v", err))
		return
	}

	if len(resources) == 0 {
		exec.Command("zenity", "--info", "--title=Resources", "--text=No resources available", "--width=300").Run()
		return
	}

	// Build list for zenity
	var options []string
	for _, res := range resources {
		label := fmt.Sprintf("%s | %s", res.Name, res.Address)
		if res.NeedsAuth {
			label += " [Locked]"
		}
		options = append(options, label)
	}

	// Show zenity list
	cmd := exec.Command("zenity", "--list", "--title=Twingate Resources",
		"--text=Available resources (select to authenticate locked resources):",
		"--column=Resource", "--width=600", "--height=400")
	cmd.Args = append(cmd.Args, options...)

	output, err := cmd.Output()
	if err != nil {
		// User cancelled
		return
	}

	selected := string(output)
	selected = selected[:len(selected)-1] // Remove trailing newline

	// Find the selected resource
	for _, res := range resources {
		if fmt.Sprintf("%s | %s", res.Name, res.Address) == selected[:len(fmt.Sprintf("%s | %s", res.Name, res.Address))] {
			if res.NeedsAuth {
				if err := twingate.AuthenticateResource(res.Name); err != nil {
					log.Printf("Failed to authenticate resource: %v", err)
					sendNotification("Authentication Failed", fmt.Sprintf("Failed to authenticate %s: %v", res.Name, err))
				} else {
					sendNotification("Authentication Started", fmt.Sprintf("Authentication initiated for %s", res.Name))
				}
			}
			break
		}
	}
}

func handleOpenWebAdmin() {
	log.Println("Opening web admin...")
	networkURL := appState.GetNetworkURL()
	log.Printf("Network URL from state: %s", networkURL)

	if networkURL == "" || networkURL == "-" {
		// Try to fetch it
		log.Println("Network URL not in state, fetching...")
		info, err := twingate.GetNetworkInfo()
		if err != nil {
			log.Printf("Failed to get network info: %v", err)
			sendNotification("Web Admin Error", fmt.Sprintf("Failed to get network info: %v", err))
			return
		}
		if info.URL == "" || info.URL == "-" {
			log.Println("Network URL not available from twingate CLI")
			sendNotification("Web Admin Error", "Network URL not available")
			return
		}
		networkURL = info.URL
	}

	// Ensure URL has proper scheme
	if !strings.HasPrefix(networkURL, "http://") && !strings.HasPrefix(networkURL, "https://") {
		networkURL = "https://" + networkURL
	}

	log.Printf("Opening URL: %s", networkURL)

	// Open URL in default browser
	cmd := exec.Command("xdg-open", networkURL)
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to open web admin: %v", err)
		sendNotification("Web Admin Error", fmt.Sprintf("Failed to open browser: %v", err))
	} else {
		log.Println("Browser opened successfully")
	}
}

func handleDiagnosticReport() {
	log.Println("Generating diagnostic report...")
	if err := twingate.GenerateDiagnosticReport(); err != nil {
		log.Printf("Failed to generate diagnostic report: %v", err)
		sendNotification("Diagnostic Report Failed", fmt.Sprintf("Failed to generate report: %v", err))
	} else {
		sendNotification("Diagnostic Report", "Diagnostic report generated successfully")
	}
}

func handleAutoConnectToggle(enabled bool) {
	log.Printf("Auto-connect toggled: %v", enabled)

	// Enable/disable the Twingate systemd service
	if err := twingate.SetAutoConnect(enabled); err != nil {
		log.Printf("Failed to set auto-connect: %v", err)
		sendNotification("Auto-connect Error", fmt.Sprintf("Failed to change auto-connect: %v", err))
		return
	}

	if enabled {
		sendNotification("Auto-connect Enabled", "Twingate service will start automatically on boot")
	} else {
		sendNotification("Auto-connect Disabled", "Twingate service will not start automatically")
	}
}

func handleMenuOpening() {
	// Update connection time immediately when menu opens
	// This ensures the displayed time is always current
	if appState.IsConnected() {
		duration := appState.GetConnectionDuration()
		if duration > 0 {
			timeStr := formatDuration(duration)
			if systemTray != nil {
				systemTray.UpdateConnectionTime(timeStr)
			}
		}
	}

	// Also refresh auto-connect status from systemd
	autoConnectEnabled := twingate.IsAutoConnectEnabled()
	if systemTray != nil {
		systemTray.SetAutoConnect(autoConnectEnabled)
	}
}

func handleQuit() {
	log.Println("Quit requested")
	cleanup()
	os.Exit(0)
}
