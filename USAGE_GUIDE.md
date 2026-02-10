# Twingate System Tray Indicator - Usage Guide

## Overview

The Twingate System Tray Indicator provides a **persistent system tray icon** for managing your Twingate VPN connection. The icon is always visible regardless of connection status, allowing you to control Twingate entirely from the system tray without any command-line operations.

## Features

✅ **Always-Visible Tray Icon** - Red when disconnected, green when connected
✅ **Background Operation** - Runs continuously in the background
✅ **Interactive Menu** - Right-click for context menu with all operations
✅ **Quick Toggle** - Left-click to toggle connect/disconnect
✅ **Status Display** - Shows current connection status in tooltip and menu
✅ **Desktop Notifications** - Notifies you when connection state changes
✅ **No CLI Required** - All operations available from tray icon

## Quick Start

### Running the Application

Simply run the binary without any arguments:

```bash
./twingate-indicator
```

Or explicitly as a daemon:

```bash
./twingate-indicator daemon
```

The application will:
1. Register with your system tray
2. Display an icon (red circle = disconnected, green circle = connected)
3. Monitor Twingate status continuously
4. Stay running in the background

### Using the Tray Icon

**Left Click** - Toggle connection
- If disconnected → Initiates connection
- If connected → Disconnects

**Right Click** - Opens context menu with options:
- **Connect** - Connect to Twingate (shown when disconnected)
- **Disconnect** - Disconnect from Twingate (shown when connected)
- **Status: Connected/Disconnected** - Current status display
- **Quit** - Exit the application

**Hover** - Shows tooltip with current connection status

## Installation

### Option 1: Manual Background Process

Run in a terminal:
```bash
./twingate-indicator &
```

To keep it running after closing the terminal:
```bash
nohup ./twingate-indicator > /tmp/twingate.log 2>&1 &
```

### Option 2: Systemd User Service (Recommended)

1. Copy the binary to a permanent location:
```bash
sudo cp twingate-indicator /usr/local/bin/
sudo chmod +x /usr/local/bin/twingate-indicator
```

2. Create systemd user service:
```bash
mkdir -p ~/.config/systemd/user/
cat > ~/.config/systemd/user/twingate-indicator.service << 'EOF'
[Unit]
Description=Twingate System Tray Indicator
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/twingate-indicator
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF
```

3. Enable and start the service:
```bash
systemctl --user daemon-reload
systemctl --user enable twingate-indicator.service
systemctl --user start twingate-indicator.service
```

4. Check status:
```bash
systemctl --user status twingate-indicator.service
```

5. View logs:
```bash
journalctl --user -u twingate-indicator.service -f
```

### Option 3: Desktop Autostart

1. Copy binary to permanent location:
```bash
sudo cp twingate-indicator /usr/local/bin/
sudo chmod +x /usr/local/bin/twingate-indicator
```

2. Create desktop entry:
```bash
mkdir -p ~/.config/autostart/
cat > ~/.config/autostart/twingate-indicator.desktop << 'EOF'
[Desktop Entry]
Type=Application
Name=Twingate Indicator
Comment=System tray indicator for Twingate VPN
Exec=/usr/local/bin/twingate-indicator
Icon=network-vpn
Terminal=false
Categories=Network;
StartupNotify=false
X-GNOME-Autostart-enabled=true
EOF
```

3. The application will start automatically on next login

## Command-Line Interface (Optional)

While the tray icon provides all functionality, CLI commands are available:

```bash
twingate-indicator status       # Check connection status
twingate-indicator connect      # Connect to Twingate
twingate-indicator disconnect   # Disconnect from Twingate
twingate-indicator help         # Show help message
```

## Technical Details

- **D-Bus Integration**: Uses StatusNotifierItem protocol for system tray
- **Desktop Compatibility**: Works with GNOME, KDE, XFCE, and other modern Linux desktops
- **Status Monitoring**: Polls Twingate status every 500ms
- **Icon Updates**: Dynamic icon changes based on connection state
- **Menu Protocol**: DBusMenu for context menu functionality
- **Binary Size**: ~6.2 MB
- **Dependencies**: godbus/dbus (bundled in binary)

## Troubleshooting

### Icon Not Visible

1. Check if the application is running:
```bash
ps aux | grep twingate-indicator
```

2. Verify D-Bus registration:
```bash
dbus-send --print-reply --session --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep twingate
```

3. Check system tray support:
```bash
dbus-send --print-reply --session --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames | grep StatusNotifierWatcher
```

### Menu Not Working

Some desktop environments may require a system tray extension. For GNOME:
```bash
gnome-extensions list | grep -i tray
```

Install AppIndicator extension if needed:
```bash
sudo apt install gnome-shell-extension-appindicator
```

### Connection Commands Fail

Ensure you have necessary permissions. The application uses `pkexec` or `sudo` to run privileged Twingate commands.

### View Application Logs

If running via systemd:
```bash
journalctl --user -u twingate-indicator.service -f
```

If running manually:
```bash
# Check where you redirected output
tail -f /tmp/twingate.log
```

## Uninstallation

### If using systemd:
```bash
systemctl --user stop twingate-indicator.service
systemctl --user disable twingate-indicator.service
rm ~/.config/systemd/user/twingate-indicator.service
systemctl --user daemon-reload
sudo rm /usr/local/bin/twingate-indicator
```

### If using autostart:
```bash
rm ~/.config/autostart/twingate-indicator.desktop
sudo rm /usr/local/bin/twingate-indicator
```

### If running manually:
```bash
killall twingate-indicator
```

## Support

For issues or feature requests, refer to the project repository or documentation.
