package twingate

import (
	"fmt"
	"strings"
)

// Resource represents a Twingate resource
type Resource struct {
	Name       string
	Address    string
	AuthStatus string
	NeedsAuth  bool
}

// GetResources returns a list of available Twingate resources
func GetResources() ([]Resource, error) {
	output, err := runCommand("twingate", "resources", "-d")
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	var resources []Resource
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		if i == 0 {
			continue // Skip header
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse TSV: Name, Address, ..., Auth Status
		fields := strings.Split(line, "\t")
		if len(fields) >= 2 {
			resource := Resource{
				Name:    strings.TrimSpace(fields[0]),
				Address: strings.TrimSpace(fields[1]),
			}

			// Auth status is typically in field 3 or 4
			if len(fields) >= 4 {
				resource.AuthStatus = strings.TrimSpace(fields[3])
			} else if len(fields) >= 3 {
				resource.AuthStatus = strings.TrimSpace(fields[2])
			}

			// Check if needs authentication
			resource.NeedsAuth = strings.Contains(strings.ToLower(resource.AuthStatus), "locked")

			if resource.Name != "" {
				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// AuthenticateResource initiates authentication for a locked resource
func AuthenticateResource(resourceName string) error {
	_, err := runCommand("twingate", "auth", resourceName)
	if err != nil {
		return fmt.Errorf("failed to authenticate resource %s: %w", resourceName, err)
	}
	return nil
}
