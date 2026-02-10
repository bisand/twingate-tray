# AGENTS.md - Twingate System Tray Indicator

Guide for AI coding agents working on this repository.

## Project Overview

A Go-based system tray indicator for Twingate VPN on Linux (GNOME and KDE Plasma).
- **Language**: Go 1.21+
- **Platform**: Linux (GNOME and KDE Plasma desktop environments)
- **Key Dependencies**: godbus/dbus v5, zenity (system package)
- **Architecture**: StatusNotifierItem D-Bus protocol with zenity popup menus

## Build Commands

### Standard Build
```bash
make build           # Build binary: twingate-tray
go build -o twingate-tray ./cmd/twingate-tray
```

### Development Workflow
```bash
make clean          # Remove binary and clean build artifacts
make fmt            # Format code with go fmt
make lint           # Run go vet for static analysis
make deps           # Tidy dependencies with go mod tidy
make all            # Clean then build
```

### Installation
```bash
make install        # Install to /usr/local/bin (requires sudo)
make uninstall      # Remove from /usr/local/bin
```

### Testing
```bash
make test           # Run all tests: go test -v ./...
go test -v          # Run tests in current package only
go test -run TestFunctionName  # Run single test (when tests exist)
```

**Note**: Currently no tests exist. When adding tests:
- Create `*_test.go` files alongside source
- Use table-driven tests for multiple cases
- Mock D-Bus interactions to avoid system dependencies

## Code Style Guidelines

### Import Ordering
Follow Go standard library conventions - alphabetically sorted with groups:
1. Standard library imports (fmt, log, os, etc.)
2. External dependencies (github.com/godbus/dbus/v5)

```go
import (
    "bytes"
    "fmt"
    "log"
    "os"
    
    "github.com/godbus/dbus/v5"
)
```

### Formatting
- **Tabs for indentation** (Go standard)
- Run `make fmt` or `go fmt ./...` before committing
- Line length: No strict limit, but keep readable (~100-120 chars preferred)
- Use `gofmt` defaults - do NOT customize

### Naming Conventions
- **Exported names**: CamelCase (SystemTray, NewSystemTray)
- **Unexported names**: camelCase (systemTray, iconPath)
- **Constants**: CamelCase (if exported) or camelCase (if private)
- **Acronyms**: Keep uppercase (DBus not Dbus, HTTP not Http)
- **Interfaces**: Suffix with -er when appropriate (Handler, Writer)

### Types and Structs
- Document all exported types with comments starting with type name
- Use struct embedding for composition
- Prefer value receivers unless you need to modify or large struct
- Group related fields in structs, use blank lines for separation

```go
// SystemTray manages the system tray icon using D-Bus StatusNotifierItem
type SystemTray struct {
    conn         *dbus.Conn
    connected    bool
    
    onConnect    func()
    onDisconnect func()
    onQuit       func()
    
    mu           sync.RWMutex
}
```

### Error Handling
- **Always check errors** - never ignore with `_` unless explicitly justified
- **Wrap errors** with context using `fmt.Errorf("context: %w", err)`
- **Log errors** before returning them when helpful for debugging
- **Return early** on errors to avoid deep nesting

```go
conn, err := dbus.SessionBus()
if err != nil {
    return nil, fmt.Errorf("failed to connect to D-Bus: %w", err)
}
```

### Concurrency
- Use `sync.RWMutex` for shared state (see SystemTray.mu)
- Lock before reading/writing shared data
- Use `defer` to unlock immediately after lock
- Launch goroutines with `go` for callbacks and async operations
- Always handle cleanup in goroutines (use defer, channels for shutdown)

```go
st.mu.RLock()
connected := st.connected
st.mu.RUnlock()
```

### Logging
- Use standard `log` package (already imported)
- Log important state changes, errors, and user actions
- Format: `log.Printf("Action: %s", value)` or `log.Println("Message")`
- Don't log in tight loops (status polling is acceptable at 500ms interval)

## Project-Specific Guidelines

### D-Bus Integration
- **StatusNotifierItem path**: `/StatusNotifierItem`
- **Menu path**: `/MenuBar`
- **Registration**: Always convert `dbus.ObjectPath` to `string` for RegisterStatusNotifierItem
- **Properties**: Implement org.freedesktop.DBus.Properties interface for compatibility
- **Cleanup**: Close D-Bus connection in Stop() method to auto-unregister

### System Tray Icon States
- **Disconnected**: `"network-wireless-offline"` (gray with X)
- **Connected**: `"network-wireless-signal-excellent"` (green signal)
- Use standard GNOME icon names, not custom pixmaps (better compatibility)

### Menu Implementation
- Use **zenity** for popup menus (more reliable than DBusMenu on GNOME)
- Launch zenity in goroutine to avoid blocking
- Parse output with `strings.TrimSpace()` for menu selections
- Handle cancellation gracefully (user closes dialog without selecting)

### Signal Handling
- Register handlers for: `SIGINT`, `SIGTERM`, `SIGQUIT`
- Always call `cleanup()` before exiting
- Use buffered channel (size 1) for signal notifications
- Launch signal handler in separate goroutine

### File Organization

The project follows standard Go project layout:

- **cmd/twingate-tray/main.go**: Application entry point, CLI handling, signal management, main loop
- **internal/app/**: Application state management and constants
  - `state.go`: Thread-safe AppState with connection status
  - `constants.go`: Status poll interval, notification timeout
- **internal/tray/**: D-Bus system tray implementation
  - `tray.go`: StatusNotifierItem + DBusMenu protocols
  - `icons.go`: Font Awesome lock/unlock icon generation
  - `constants.go`: Menu item IDs, icon specifications
- **internal/twingate/**: VPN interaction
  - `cli.go`: Twingate CLI wrapper and privilege escalation
  - `status.go`: Connection info gathering, dialogs, clipboard
- **tools/**: Build and development utilities
  - `generate_icons.go`: PNG icon generator (build tool)
  - `svg_to_points.py`: SVG to polygon converter

Keep files focused - create new packages if adding major features

## Common Pitfalls to Avoid

1. **Don't** use D-Bus object paths directly as strings without conversion
2. **Don't** block the main thread with menu operations (use goroutines)
3. **Don't** forget to update icon when connection state changes
4. **Don't** use DBusMenu (complex, poor GNOME compatibility) - use zenity instead
5. **Don't** create multiple instances - kill existing before starting new one
6. **Don't** log excessively in status monitoring loop
7. **Don't** assume twingate CLI exists - handle errors gracefully
8. **Don't** forget mutex locks when accessing SystemTray.connected

## Testing Checklist

When making changes, manually verify:
- [ ] Icon appears in system tray after starting
- [ ] Icon changes when VPN connects/disconnects
- [ ] Click shows popup menu with correct options
- [ ] Menu selections execute correct actions
- [ ] Icon disappears when app exits (Ctrl+C or Quit)
- [ ] No stale icons remain after crash (test with `kill -9`)
- [ ] Logs show proper registration: "Registered with StatusNotifierWatcher"
- [ ] No D-Bus errors in logs

## Dependencies Management

- **godbus/dbus v5.1.0**: D-Bus bindings (required)
- **zenity**: System package for popup dialogs (verify with `which zenity`)
- **twingate**: CLI tool (user must install separately)

Update dependencies:
```bash
go get github.com/godbus/dbus/v5@latest
go mod tidy
```

## Release Process

1. Update version if tracking in code
2. Run `make clean && make all`
3. Test on target system (GNOME desktop environment)
4. Tag release: `git tag v1.0.0`
5. Build binary: `make build`
6. Create release package with install.sh script

## Additional Notes

- This is a **Linux-only** application (uses D-Bus, GNOME icons, zenity)
- StatusNotifierItem is the modern Linux tray standard (native to KDE, requires extension on GNOME)
- GNOME Shell AppIndicator extension must be enabled on GNOME systems
- KDE Plasma has native StatusNotifierItem support
- Privilege escalation uses `pkexec` (primary) or `sudo` (fallback) for VPN operations
