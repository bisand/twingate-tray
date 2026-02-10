#!/bin/bash
# Twingate Tray - Start Script

BINARY="/home/bisand/dev/twingate-tray/twingate-tray"

# Kill any existing instances
killall -q twingate-tray 2>/dev/null

# Start in background
nohup "$BINARY" >/tmp/twingate-tray.log 2>&1 &

echo "Twingate Tray started!"
echo "Check your system tray for the icon"
echo ""
echo "Logs: tail -f /tmp/twingate-tray.log"
echo "Stop: killall twingate-tray"
