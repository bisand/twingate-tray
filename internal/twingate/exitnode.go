package twingate

import (
	"fmt"
	"strings"
)

// ExitNodeStatus represents the exit node status
type ExitNodeStatus struct {
	Enabled        bool
	CurrentNode    string
	AvailableNodes []string
}

// GetExitNodeStatus returns current exit node status
func GetExitNodeStatus() (*ExitNodeStatus, error) {
	status := &ExitNodeStatus{}

	// Check if exit node routing is active by listing nodes
	output, err := runCommand("twingate", "exit-node", "list", "-d")
	if err != nil {
		return nil, fmt.Errorf("failed to get exit node list: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Name") {
			continue // Skip header
		}

		// Parse TSV: Name, Location, Active
		fields := strings.Split(line, "\t")
		if len(fields) >= 1 {
			nodeName := strings.TrimSpace(fields[0])
			status.AvailableNodes = append(status.AvailableNodes, nodeName)

			// Check if this is the active node
			if len(fields) >= 3 && strings.TrimSpace(fields[2]) == "true" {
				status.Enabled = true
				status.CurrentNode = nodeName
			}
		}
	}

	return status, nil
}

// StartExitNode starts routing all traffic through Twingate
func StartExitNode() error {
	if err := runPrivilegedCommand("twingate", "exit-node", "start"); err != nil {
		return fmt.Errorf("failed to start exit node: %w", err)
	}
	return nil
}

// StopExitNode stops routing all traffic through Twingate
func StopExitNode() error {
	if err := runPrivilegedCommand("twingate", "exit-node", "stop"); err != nil {
		return fmt.Errorf("failed to stop exit node: %w", err)
	}
	return nil
}

// SwitchExitNode switches to a different exit node
func SwitchExitNode(nodeName string) error {
	if err := runPrivilegedCommand("twingate", "exit-node", "switch", nodeName); err != nil {
		return fmt.Errorf("failed to switch exit node to %s: %w", nodeName, err)
	}
	return nil
}
