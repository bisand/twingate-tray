#!/bin/bash
# Twingate Indicator - Start Script

BINARY="/home/bisand/dev/twingate-ui/twingate-indicator"

# Kill any existing instances
killall -q twingate-indicator 2>/dev/null

# Start in background
nohup "$BINARY" >/tmp/twingate-indicator.log 2>&1 &

echo "Twingate Indicator started!"
echo "Check your system tray for the icon"
echo ""
echo "Logs: tail -f /tmp/twingate-indicator.log"
echo "Stop: killall twingate-indicator"
