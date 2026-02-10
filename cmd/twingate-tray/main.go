package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
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

	// Initialize system tray
	var err error
	systemTray, err = tray.NewSystemTray(
		handleConnect,
		handleDisconnect,
		handleConnectionInfo,
		handleQuit,
	)

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
		connected, err := twingate.CheckStatus()

		appState.SetConnected(connected)

		if err != nil {
			appState.SetLastError(err.Error())
		} else {
			appState.SetLastError("")
		}

		// Update system tray icon
		if systemTray != nil {
			systemTray.UpdateStatus(connected)
		}

		// Log status changes
		if prevConnected == nil || *prevConnected != connected {
			if connected {
				log.Println("Status: Connected to Twingate")
				sendNotification("Twingate Connected", "You are now connected to Twingate")
			} else {
				log.Println("Status: Disconnected from Twingate")
				sendNotification("Twingate Disconnected", "You are now disconnected from Twingate")
			}
			prevConnected = &connected
		}

		// Log errors
		if err != nil {
			log.Printf("Status check error: %v", err)
		}
	}
}

func sendNotification(title, body string) {
	// Try to send via notify-send if available
	cmd := exec.Command("notify-send", "-t", fmt.Sprintf("%d", app.NotificationTimeout), title, body)
	// Ignore errors - notification is optional
	_ = cmd.Run()
}

// Handler functions for tray callbacks

func handleConnect() {
	if err := twingate.Connect(); err != nil {
		log.Printf("Connect failed: %v", err)
	}
}

func handleDisconnect() {
	if err := twingate.Disconnect(); err != nil {
		log.Printf("Disconnect failed: %v", err)
	}
}

func handleConnectionInfo() {
	twingate.ShowConnectionInfo()
}

func handleQuit() {
	log.Println("Quit requested")
	cleanup()
	os.Exit(0)
}
