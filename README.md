# Twingate System Tray Indicator

A lightweight system tray indicator for Twingate VPN on Linux (GNOME and KDE). Displays real-time connection status with a native context menu for managing VPN connections.

## Features

- **System Tray Integration**: Native StatusNotifierItem (SNI) tray icon via D-Bus
- **Visual Status Indicator**: Lock icons showing connected (locked) / disconnected (unlocked) states
- **Native Context Menu**: Right-click menu with Connect/Disconnect, Connection Info, and Quit options
- **Connection Info Dialog**: Detailed connection information with copy-to-clipboard functionality
- **Desktop Notifications**: System notifications on connection status changes
- **Privilege Escalation**: Supports both `pkexec` and `sudo` for secure connection management
- **Native Clipboard Support**: Built-in clipboard integration via X11 (no external tools needed)

## Installation

### Quick Install (Automated)

Use the provided install script:

```bash
./install.sh
```

This will:
1. Build the binary
2. Install to `/usr/local/bin/`
3. Set up autostart (optional)

### Build from Source

```bash
# Install build dependencies (Ubuntu/Debian)
sudo apt install golang-go libx11-dev build-essential

# Clone or navigate to the project directory
cd twingate-tray

# Build the binary using make
make build

# Or build directly with go
go build -o twingate-tray ./cmd/twingate-tray

# Optional: Install to system path
sudo make install
# Or manually:
sudo cp twingate-tray /usr/local/bin/
```

### Manual Start

```bash
# Start the indicator
./twingate-tray

# Or if installed to /usr/local/bin:
twingate-tray
```

## Usage

### System Tray

Once started, the indicator appears in your system tray:

- **Icon States**:
  - 🔒 Locked (white) = Connected to Twingate
  - 🔓 Unlocked (gray) = Disconnected

- **Menu Actions**:
  - **Connect/Disconnect**: Toggle VPN connection (requires sudo/pkexec)
  - **Connection Info...**: View detailed connection information
    - Shows: Status, IP addresses, DNS, routes, resources, daemon info
    - **Copy to Clipboard** button: Copy all info as plain text
  - **Quit**: Exit the indicator

### CLI Mode

The application also supports command-line mode:

```bash
# Check current connection status
twingate-tray status
# Output: "connected" or "disconnected"

# Connect to Twingate (requires elevated privileges)
twingate-tray connect

# Disconnect from Twingate (requires elevated privileges)
twingate-tray disconnect

# Show help
twingate-tray help
```

Default behavior (no arguments) launches the system tray.

### Run as System Service

Create a systemd user service file at `~/.config/systemd/user/twingate-tray.service`:

```ini
[Unit]
Description=Twingate System Tray Indicator
After=graphical-session.target

[Service]
Type=simple
ExecStart=%h/.local/bin/twingate-tray
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

Then enable and start it:

```bash
systemctl --user enable twingate-tray
systemctl --user start twingate-tray
systemctl --user status twingate-tray
```

View logs:

```bash
journalctl --user -u twingate-tray -f
```

## How It Works

### System Tray Architecture

1. **D-Bus Integration**: Registers as a StatusNotifierItem on the D-Bus session bus
2. **StatusNotifierWatcher**: Communicates with the system's status notifier watcher
3. **DBusMenu Protocol**: Provides native context menu via `com.canonical.dbusmenu`
4. **Icon Rendering**: Generates lock/unlock icons dynamically using polygon rasterization
5. **Status Monitoring**: Polls `twingate status` every 500ms to detect connection changes

### Connection Info Dialog

- **Data Sources**: Gathers information from multiple commands:
  - `twingate status -v -d`: Connection status and Secure DNS
  - `twingate account list -d`: User email, network name
  - `twingate resources -d`: Available resources
  - `ip addr show sdwan0`: Network interface details
  - `resolvectl status sdwan0`: DNS configuration
  - `systemctl show twingate`: Daemon uptime and memory
- **Clipboard**: Uses native X11 clipboard API (golang.design/x/clipboard)
- **Dialog**: zenity `--text-info` for scrollable, selectable text view

## Requirements

### Runtime Requirements
- **Twingate CLI**: Must be installed and in PATH
- **D-Bus Session Bus**: For system tray communication (standard on all Linux desktops)
- **System Tray Support**: 
  - GNOME: AppIndicator/KStatusNotifierItem extension required
  - KDE Plasma: Native support (works out of the box)
- **zenity**: For popup dialogs (usually pre-installed)
- **notify-send**: For desktop notifications (optional, from `libnotify-bin` package)
- **pkexec or sudo**: For privileged VPN operations

### Build Dependencies
To build from source, you need:
- **Go 1.21+**: Go compiler and toolchain
- **libx11-dev**: X11 development headers (for native clipboard support via CGO)
- **make**: Build automation (optional, can use `go build` directly)

Install build dependencies on Ubuntu/Debian:
```bash
sudo apt install golang-go libx11-dev build-essential
```

## Architecture

```
twingate-tray/
├── cmd/
│   └── twingate-tray/
│       └── main.go           # Application entry point, CLI handling
├── internal/
│   ├── app/
│   │   ├── state.go          # Application state management
│   │   └── constants.go      # Application-level constants
│   ├── tray/
│   │   ├── tray.go           # D-Bus system tray (StatusNotifierItem + DBusMenu)
│   │   ├── icons.go          # Lock/unlock icon generation (Font Awesome)
│   │   └── constants.go      # Menu item IDs and icon specs
│   └── twingate/
│       ├── cli.go            # Twingate CLI wrapper and privilege escalation
│       └── status.go         # Connection info gathering and dialogs
├── tools/
│   ├── generate_icons.go     # PNG icon generator (build tool)
│   └── svg_to_points.py      # SVG to polygon converter
├── assets/                   # Static icon files
├── go.mod                    # Go module definition with dependencies
├── Makefile                  # Build automation
├── install.sh                # Installation script
└── README.md                 # This file
```

### Key Components

- **SystemTray**: D-Bus tray implementation with SNI and DBusMenu protocols
- **ConnectionInfo**: Aggregates VPN status from multiple sources
- **Icon Generation**: Scanline rasterizer for Font Awesome lock icons
- **Privilege Escalation**: pkexec → sudo fallback for VPN control
- **Clipboard Integration**: Native X11 clipboard via CGO

## Development

### Build Commands

```bash
make build           # Build binary: twingate-tray
make clean          # Remove binary and clean build artifacts
make fmt            # Format code with go fmt
make lint           # Run go vet for static analysis
make deps           # Tidy dependencies with go mod tidy
make all            # Clean then build
make install        # Install to /usr/local/bin (requires sudo)
make uninstall      # Remove from /usr/local/bin
make test           # Run tests (when available)
```

### Dependencies

The project uses:
- **github.com/godbus/dbus/v5**: D-Bus bindings for Go
- **golang.design/x/clipboard**: Native clipboard support (requires libx11-dev)

Update dependencies:
```bash
go get -u ./...
go mod tidy
```

### Testing Manually

After building, test these scenarios:
- [ ] Icon appears in system tray
- [ ] Icon changes between locked/unlocked states
- [ ] Right-click shows menu with all options
- [ ] Connect/Disconnect executes correctly
- [ ] Connection Info dialog shows all fields
- [ ] Copy to Clipboard works
- [ ] Quit removes icon cleanly

## Troubleshooting

### System Tray Icon Not Appearing

- **GNOME**: Ensure AppIndicator extension is installed and enabled
  ```bash
  gnome-extensions list | grep appindicator
  ```
- **Check D-Bus registration**:
  ```bash
  # Should show the indicator's bus name
  busctl --user list | grep StatusNotifierItem
  ```
- **View logs**: Run in foreground to see errors
  ```bash
  ./twingate-tray
  ```

### Connection Status Not Updating

- Verify Twingate CLI is installed: `which twingate`
- Check Twingate status manually: `twingate status`
- Ensure Twingate daemon is running: `systemctl status twingate`

### Privilege Escalation Fails

- Ensure you have either `pkexec` or `sudo` installed
- Check that your user has appropriate permissions
- Try running manually: `pkexec twingate start` or `sudo twingate start`

### Clipboard Copy Not Working

- **Build issue**: Ensure `libx11-dev` was installed before building
  ```bash
  # Rebuild if you installed libx11-dev after initial build
  make clean && make build
  ```
- **Runtime check**: The clipboard uses X11, verify you're on X11 (not Wayland-only)
  ```bash
  echo $XDG_SESSION_TYPE  # Should show "x11" or "wayland"
  ```

### Connection Info Dialog Shows Partial Data

- Some fields require specific commands to be available
- Check if these are in your PATH: `ip`, `resolvectl`, `systemctl`
- Missing data appears as "-" in the dialog

## Platform Support

### Tested On
- ✅ GNOME 40+ (Ubuntu, Fedora, Debian)
- ✅ KDE Plasma 5.20+

### Requirements by Desktop Environment
- **GNOME**: AppIndicator extension required (available in most distributions' extension repositories)
- **KDE Plasma**: Works natively out of the box (StatusNotifierItem is native to KDE)
- **Other DEs**: May work if they support StatusNotifierItem/AppIndicator protocol

### Known Limitations

- **Linux only**: Uses Linux-specific D-Bus protocols and system commands
- **X11/Wayland**: Clipboard requires X11 libraries (works on Wayland via XWayland)
- **SystemD**: Connection info gathering assumes systemd for daemon stats
- **Twingate interface name**: Hardcoded to `sdwan0` (Twingate's default)

## Technical Details

### D-Bus Protocols Used
- **org.kde.StatusNotifierItem**: System tray icon protocol
- **com.canonical.dbusmenu**: Native context menu protocol
- **org.freedesktop.DBus.Properties**: D-Bus properties interface
- **org.kde.StatusNotifierWatcher**: System tray registration

### Icon Rendering
- Font Awesome lock/unlock icons
- Polygon-based scanline rasterization
- 2x supersampled anti-aliasing
- 256x256 ARGB pixel format for D-Bus IconPixmap

### Clipboard Implementation
- Uses `golang.design/x/clipboard` library
- CGO bindings to X11 clipboard API
- Supports both X11 and Wayland (via XWayland)
- No external clipboard tools (xclip/xsel) required

## License

MIT License (or specify your license)

## Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly on GNOME/KDE
5. Submit a pull request
