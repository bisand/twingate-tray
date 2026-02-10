package app

import (
	"sync"
	"time"
)

// AppState tracks the current Twingate connection status
type AppState struct {
	mu             sync.RWMutex
	connected      bool
	lastErr        string
	networkName    string
	networkURL     string
	connectedSince time.Time
}

// NewAppState creates a new application state
func NewAppState() *AppState {
	return &AppState{
		connected:   false,
		lastErr:     "",
		networkName: "",
		networkURL:  "",
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
	if connected && !a.connected {
		// Just connected - record the time
		a.connectedSince = time.Now()
	} else if !connected && a.connected {
		// Just disconnected - reset the time
		a.connectedSince = time.Time{}
	}
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

// GetNetworkName returns the current network name
func (a *AppState) GetNetworkName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.networkName
}

// SetNetworkName updates the network name
func (a *AppState) SetNetworkName(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.networkName = name
}

// GetNetworkURL returns the current network URL
func (a *AppState) GetNetworkURL() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.networkURL
}

// SetNetworkURL updates the network URL
func (a *AppState) SetNetworkURL(url string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.networkURL = url
}

// GetConnectedSince returns the time when connection was established
func (a *AppState) GetConnectedSince() time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.connectedSince
}

// GetConnectionDuration returns how long we've been connected
func (a *AppState) GetConnectionDuration() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if !a.connected || a.connectedSince.IsZero() {
		return 0
	}
	return time.Since(a.connectedSince)
}
