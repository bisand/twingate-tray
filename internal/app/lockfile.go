package app

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// LockFile provides single-instance guard functionality
type LockFile struct {
	path string
}

// NewLockFile creates a new LockFile instance
func NewLockFile() *LockFile {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = filepath.Join(os.TempDir(), "user-"+getCurrentUID())
	}

	// Ensure directory exists
	_ = os.MkdirAll(runtimeDir, 0755)

	lockPath := filepath.Join(runtimeDir, "twingate-tray.lock")
	return &LockFile{path: lockPath}
}

// Acquire attempts to acquire the lock, returning an error if another instance exists
func (lf *LockFile) Acquire() error {
	// Check if lock file exists and if the PID in it is still running
	if data, err := os.ReadFile(lf.path); err == nil {
		oldPID := strings.TrimSpace(string(data))
		if isProcessRunning(oldPID) {
			return fmt.Errorf("another instance is already running (PID: %s)", oldPID)
		}
		// Old process is not running, clean up stale lock file
		_ = os.Remove(lf.path)
	}

	// Write current PID to lock file
	currentPID := os.Getpid()
	pidStr := strconv.Itoa(currentPID)

	if err := os.WriteFile(lf.path, []byte(pidStr+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	log.Printf("Lock acquired (PID: %d)", currentPID)
	return nil
}

// Release removes the lock file
func (lf *LockFile) Release() {
	if err := os.Remove(lf.path); err == nil {
		log.Println("Lock released")
	}
}

// isProcessRunning checks if a process with the given PID is still running
func isProcessRunning(pidStr string) bool {
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	// Use syscall.Kill with signal 0 to check if the process exists
	// Signal 0 checks for process existence without sending a signal
	err = syscall.Kill(pid, 0)
	return err == nil
}

// getCurrentUID returns the current user's UID as a string
func getCurrentUID() string {
	u, err := user.Current()
	if err != nil {
		return "user"
	}
	return u.Uid
}
