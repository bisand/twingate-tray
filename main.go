package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// AppState tracks the current twingate connection status
type AppState struct {
	mu        sync.RWMutex
	connected bool
	lastErr   string
}

var appState AppState
var systemTray *SystemTray

func main() {
	if len(os.Args) > 1 {
		// CLI mode - handle commands
		handleCLI(os.Args[1:])
		return
	}

	// Daemon mode with system tray
	log.Println("Starting Twingate indicator with system tray...")

	appState.connected = false
	appState.lastErr = ""

	// Initialize system tray
	var err error
	systemTray, err = NewSystemTray(func() {
		if e := handleConnect(); e != nil {
			log.Printf("Connect failed: %v", e)
		}
	}, func() {
		if e := handleDisconnect(); e != nil {
			log.Printf("Disconnect failed: %v", e)
		}
	}, func() {
		log.Println("Quit requested")
		cleanup()
		os.Exit(0)
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
		connected, err := checkTwingateStatus()
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
		if err := handleConnect(); err != nil {
			fmt.Printf("Error connecting: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Connection initiated")

	case "disconnect":
		if err := handleDisconnect(); err != nil {
			fmt.Printf("Error disconnecting: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Disconnection initiated")

	case "daemon":
		// Start as daemon with system tray
		log.Println("Starting Twingate indicator daemon...")
		appState.connected = false
		appState.lastErr = ""

		// Initialize system tray
		var err error
		systemTray, err = NewSystemTray(func() {
			if e := handleConnect(); e != nil {
				log.Printf("Connect failed: %v", e)
			}
		}, func() {
			if e := handleDisconnect(); e != nil {
				log.Printf("Disconnect failed: %v", e)
			}
		}, func() {
			log.Println("Quit requested")
			cleanup()
			os.Exit(0)
		})
		if err != nil {
			log.Printf("Warning: Could not initialize system tray: %v", err)
			os.Exit(1)
		}

		err = systemTray.Start()
		if err != nil {
			log.Printf("Warning: Could not start system tray: %v", err)
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

	case "help", "-h", "--help":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Twingate Indicator - System tray indicator for Twingate

Usage:
  twingate-indicator                    # Run with system tray
  twingate-indicator status             # Check connection status
  twingate-indicator connect            # Connect to Twingate
  twingate-indicator disconnect         # Disconnect from Twingate
  twingate-indicator daemon             # Start as daemon with system tray
  twingate-indicator help               # Show this help message`)
}

func monitorStatus() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var prevConnected *bool

	for range ticker.C {
		connected, err := checkTwingateStatus()

		appState.mu.Lock()
		appState.connected = connected

		if err != nil {
			appState.lastErr = err.Error()
		} else {
			appState.lastErr = ""
		}
		appState.mu.Unlock()

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
	cmd := exec.Command("notify-send", "-t", "5000", title, body)
	// Ignore errors - notification is optional
	_ = cmd.Run()
}
