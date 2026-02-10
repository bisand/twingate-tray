package app

import "time"

// Application constants
const (
	// StatusPollInterval is how often we check Twingate status
	StatusPollInterval = 500 * time.Millisecond

	// NotificationTimeout is the default notification display duration (milliseconds)
	NotificationTimeout = 5000
)
