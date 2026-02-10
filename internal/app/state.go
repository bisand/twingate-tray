package app

import (
	"sync"
)

// AppState tracks the current Twingate connection status
type AppState struct {
	mu        sync.RWMutex
	connected bool
	lastErr   string
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		connected: false,
		lastErr:   "",
	}
}

// IsConnected returns the current connection status
func (a *AppState) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connected
}

// SetConnected updates the connection status
func (a *AppState) SetConnected(connected bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = connected
}

// GetLastError returns the last error message
func (a *AppState) GetLastError() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.lastErr
}

// SetLastError updates the last error message
func (a *AppState) SetLastError(err string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastErr = err
}
