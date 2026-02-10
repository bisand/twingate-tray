#!/bin/bash
# Twingate Tray - Installation Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/twingate-tray"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="$HOME/.config/systemd/user"
AUTOSTART_DIR="$HOME/.config/autostart"

echo "========================================="
echo "Twingate Tray - Installation"
echo "========================================="
echo ""

# Check if binary exists
if [ ! -f "$BINARY" ]; then
	echo "Error: Binary not found at $BINARY"
	echo "Please run 'make build' first"
	exit 1
fi

# Make binary executable
chmod +x "$BINARY"

echo "Choose installation method:"
echo "  1) Systemd user service (recommended - auto-restart on failure)"
echo "  2) Desktop autostart (starts on login)"
echo "  3) Copy to $INSTALL_DIR only (manual management)"
echo "  4) Run in current terminal (temporary)"
echo ""
read -p "Enter choice [1-4]: " choice

case $choice in
1)
	echo ""
	echo "Installing as systemd user service..."

	# Copy binary
	sudo cp "$BINARY" "$INSTALL_DIR/twingate-tray"
	sudo chmod +x "$INSTALL_DIR/twingate-tray"
	echo "✓ Binary copied to $INSTALL_DIR/twingate-tray"

	# Create service file
	mkdir -p "$SERVICE_DIR"
	cat >"$SERVICE_DIR/twingate-tray.service" <<'EOF'
[Unit]
Description=Twingate System Tray Indicator
After=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/twingate-tray
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF
	echo "✓ Service file created"

	# Enable and start
	systemctl --user daemon-reload
	systemctl --user enable twingate-tray.service
	systemctl --user start twingate-tray.service

	echo ""
	echo "✓ Installation complete!"
	echo ""
	echo "Service status:"
	systemctl --user status twingate-tray.service --no-pager -l
	echo ""
	echo "Useful commands:"
	echo "  systemctl --user status twingate-tray.service"
	echo "  systemctl --user restart twingate-tray.service"
	echo "  systemctl --user stop twingate-tray.service"
	echo "  journalctl --user -u twingate-tray.service -f"
	;;

2)
	echo ""
	echo "Installing as desktop autostart..."

	# Copy binary
	sudo cp "$BINARY" "$INSTALL_DIR/twingate-tray"
	sudo chmod +x "$INSTALL_DIR/twingate-tray"
	echo "✓ Binary copied to $INSTALL_DIR/twingate-tray"

	# Create desktop entry
	mkdir -p "$AUTOSTART_DIR"
	cat >"$AUTOSTART_DIR/twingate-tray.desktop" <<'EOF'
[Desktop Entry]
Type=Application
Name=Twingate Tray
Comment=System tray indicator for Twingate VPN
Exec=/usr/local/bin/twingate-tray
Icon=network-vpn
Terminal=false
Categories=Network;
StartupNotify=false
X-GNOME-Autostart-enabled=true
EOF
	echo "✓ Autostart entry created"

	# Start it now
	/usr/local/bin/twingate-tray &
	sleep 2

	echo ""
	echo "✓ Installation complete!"
	echo ""
	echo "The indicator is now running and will start automatically on login."
	echo ""
	echo "To stop: killall twingate-tray"
	echo "To disable autostart: rm $AUTOSTART_DIR/twingate-tray.desktop"
	;;

3)
	echo ""
	echo "Copying binary to $INSTALL_DIR..."
	sudo cp "$BINARY" "$INSTALL_DIR/twingate-tray"
	sudo chmod +x "$INSTALL_DIR/twingate-tray"

	echo ""
	echo "✓ Installation complete!"
	echo ""
	echo "Run manually:"
	echo "  twingate-tray"
	echo ""
	echo "Run in background:"
	echo "  nohup twingate-tray > /tmp/twingate.log 2>&1 &"
	;;

4)
	echo ""
	echo "Starting in current terminal..."
	echo "Press Ctrl+C to stop"
	echo ""
	sleep 2
	"$BINARY"
	;;

*)
	echo "Invalid choice"
	exit 1
	;;
esac

echo ""
echo "========================================="
echo "Look for the tray icon in the top-right"
echo "corner of your screen!"
echo "========================================="
