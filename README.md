# Twingate Indicator

A lightweight, single-binary Go application that provides a system tray indicator and CLI interface for managing Twingate VPN connections on Linux.

## Features

- **Daemon Mode**: Run as a background service that monitors Twingate connection status
- **CLI Interface**: Simple command-line interface for checking and controlling Twingate
- **Desktop Notifications**: Optional system notifications on connection status changes (via `notify-send`)
- **Zero Dependencies**: Single compiled binary with no external dependencies (except Twingate CLI itself)
- **Privilege Escalation**: Supports both `pkexec` and `sudo` for secure connection management
- **Cross-Distribution**: Works on any systemd-based Linux distribution

## Installation

### Build from Source

```bash
# Clone or navigate to the project directory
cd /home/bisand/dev/twingate-ui

# Build the binary
go build -o twingate-indicator .

# Optional: Install to a system path
sudo cp twingate-indicator /usr/local/bin/
```

### Using the Binary

The compiled binary is ready to use immediately:

```bash
./twingate-indicator --help
```

## Usage

### CLI Commands

```bash
# Run as daemon (monitors status and sends notifications)
twingate-indicator

# Or explicitly:
twingate-indicator daemon

# Check current connection status
twingate-indicator status
# Output: "connected" or "disconnected"

# Connect to Twingate (requires elevated privileges)
twingate-indicator connect

# Disconnect from Twingate (requires elevated privileges)
twingate-indicator disconnect

# Show help
twingate-indicator help
```

### Run as System Service

Create a systemd user service file at `~/.config/systemd/user/twingate-indicator.service`:

```ini
[Unit]
Description=Twingate Indicator
After=network-online.target

[Service]
Type=simple
ExecStart=%h/.local/bin/twingate-indicator
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

Then enable and start it:

```bash
systemctl --user enable twingate-indicator
systemctl --user start twingate-indicator
systemctl --user status twingate-indicator
```

View logs:

```bash
journalctl --user -u twingate-indicator -f
```

## How It Works

1. **Status Monitoring**: The daemon polls `twingate status` every 500ms to check connection state
2. **State Changes**: When connection status changes, it:
   - Logs the change
   - Sends a system notification (if `notify-send` is available)
3. **CLI Actions**: 
   - `connect` and `disconnect` commands execute privileged Twingate commands
   - Attempts privilege escalation via `pkexec` first, falls back to `sudo`
   - Also runs `twingate desktop-restart` or `desktop-stop` for proper desktop integration

## Requirements

- Twingate CLI installed and in PATH
- `notify-send` (optional, for desktop notifications)
- For privileged operations: `pkexec` or `sudo`

## Architecture

```
twingate-indicator/
├── main.go              # Main application, daemon, and CLI handling
├── twingate.go          # Twingate CLI wrapper and command execution
├── go.mod               # Go module definition
└── README.md            # This file
```

### Key Components

- **AppState**: Thread-safe state tracking for connection status
- **monitorStatus()**: Goroutine that polls Twingate status every 500ms
- **handleCLI()**: Command router for CLI mode
- **runPrivilegedCommand()**: Privilege escalation helper (pkexec → sudo fallback)
- **sendNotification()**: Desktop notification sender

## Development

### Rebuild After Changes

```bash
go build -o twingate-indicator .
```

### Test Commands

```bash
# Test without installing (assuming Twingate is not available)
./twingate-indicator status
./twingate-indicator help
```

## Troubleshooting

### Connection Status Always Shows "disconnected"

- Verify Twingate CLI is installed: `which twingate`
- Check Twingate status manually: `twingate status`
- Review daemon logs: `journalctl --user -u twingate-indicator` (if running as service)

### Privilege Escalation Fails

- Ensure you have either `pkexec` or `sudo` installed
- Check that your user has appropriate permissions
- Try running manually with `pkexec twingate start` or `sudo twingate start`

### Notifications Not Appearing

- Install `libnotify-bin`: `sudo apt-get install libnotify-bin`
- Or disable notifications by not running the daemon (use CLI mode instead)

## Limitations

- Requires Twingate CLI to be installed separately
- Desktop notifications require `notify-send` (Ubuntu/Debian: `libnotify-bin` package)
- Privilege escalation requires appropriate system permissions
- System tray integration is currently via CLI/daemon mode (no traditional tray icon in this version)

## Future Enhancements

Potential improvements for future versions:

- GTK/D-Bus system tray icon integration (platform-specific)
- Systemd user timer for status checks
- Configuration file support
- Resource listing and browser integration
- Wayland clipboard support

## License

(Add your license here)

## Contributing

(Add contribution guidelines here)
# twingate-tray
