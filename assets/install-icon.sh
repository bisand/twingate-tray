#!/bin/bash
# Install Twingate Tray icon to system paths

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ICON_DIR="/usr/share/icons/hicolor"
APP_DIR="/usr/share/applications"
SVG_FILE="$SCRIPT_DIR/twingate-tray.svg"

echo "Installing Twingate Tray icon..."

# Install SVG icon (scalable)
sudo mkdir -p "$ICON_DIR/scalable/apps"
sudo cp "$SVG_FILE" "$ICON_DIR/scalable/apps/twingate-tray.svg"
echo "✓ Installed SVG icon"

# Install PNG icons at various sizes if conversion tools are available
if command -v rsvg-convert &>/dev/null || command -v inkscape &>/dev/null || command -v convert &>/dev/null; then
	for size in 16 22 24 32 48 64 128 256 512; do
		SIZE_DIR="$ICON_DIR/${size}x${size}/apps"
		sudo mkdir -p "$SIZE_DIR"

		# Convert SVG to PNG at this size
		if command -v rsvg-convert &>/dev/null; then
			sudo rsvg-convert -w $size -h $size "$SVG_FILE" -o "$SIZE_DIR/twingate-tray.png"
		elif command -v inkscape &>/dev/null; then
			inkscape "$SVG_FILE" --export-type=png --export-filename="$SIZE_DIR/twingate-tray.png" -w $size -h $size 2>/dev/null
		elif command -v convert &>/dev/null; then
			convert -background none -resize ${size}x${size} "$SVG_FILE" "$SIZE_DIR/twingate-tray.png"
		fi

		if [ -f "$SIZE_DIR/twingate-tray.png" ]; then
			echo "✓ Installed ${size}x${size} PNG icon"
		fi
	done
else
	echo "⚠ No SVG converter found (rsvg-convert, inkscape, or imagemagick)"
	echo "  Only SVG icon installed. Install one of these tools for PNG icons."
fi

# Install desktop file
if [ -f "$SCRIPT_DIR/twingate-tray.desktop" ]; then
	sudo cp "$SCRIPT_DIR/twingate-tray.desktop" "$APP_DIR/"
	echo "✓ Installed desktop file"
fi

# Update icon cache
if command -v gtk-update-icon-cache &>/dev/null; then
	sudo gtk-update-icon-cache -f -t "$ICON_DIR" 2>/dev/null || true
	echo "✓ Updated icon cache"
fi

echo ""
echo "Icon installation complete!"
echo "The icon is now available system-wide as 'twingate-tray'"
