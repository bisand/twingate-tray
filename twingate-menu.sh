#!/bin/bash
# Twingate Indicator Menu Helper
# This script shows a context menu for the tray icon

INDICATOR_BIN="$(dirname "$0")/twingate-indicator"

# Check current status
STATUS=$("$INDICATOR_BIN" status)

# Try to show menu with available tools
if command -v zenity &>/dev/null; then
	# GNOME/GTK dialog
	CHOICE=$(zenity --list \
		--title="Twingate Indicator" \
		--text="Status: $STATUS" \
		--column="Option" \
		$([ "$STATUS" = "disconnected" ] && echo "Connect" || echo "Disconnect") \
		"Quit")

	case "$CHOICE" in
	"Connect")
		"$INDICATOR_BIN" connect
		;;
	"Disconnect")
		"$INDICATOR_BIN" disconnect
		;;
	"Quit")
		exit 0
		;;
	esac
elif command -v kdialog &>/dev/null; then
	# KDE dialog
	kdialog --sorry "Twingate Status: $STATUS"
else
	# Fallback to notify-send
	notify-send "Twingate Status" "Status: $STATUS"
fi
