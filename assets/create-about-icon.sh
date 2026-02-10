#!/bin/bash
# Create 96x96 PNG icon for About dialog from SVG

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SVG_FILE="${SCRIPT_DIR}/twingate-tray.svg"
PNG_FILE="${SCRIPT_DIR}/twingate-tray-96.png"

# Check if SVG exists
if [ ! -f "$SVG_FILE" ]; then
	echo "Error: $SVG_FILE not found"
	exit 1
fi

# Try different conversion tools in order of preference
if command -v rsvg-convert &>/dev/null; then
	echo "Creating About dialog icon with rsvg-convert..."
	rsvg-convert -w 96 -h 96 "$SVG_FILE" -o "$PNG_FILE"
	echo "Created: $PNG_FILE (96x96)"
elif command -v convert &>/dev/null; then
	echo "Creating About dialog icon with ImageMagick convert..."
	convert -background none -resize 96x96 "$SVG_FILE" "$PNG_FILE"
	echo "Created: $PNG_FILE (96x96)"
elif command -v magick &>/dev/null; then
	echo "Creating About dialog icon with ImageMagick magick..."
	magick -background none -resize 96x96 "$SVG_FILE" "$PNG_FILE"
	echo "Created: $PNG_FILE (96x96)"
elif command -v inkscape &>/dev/null; then
	echo "Creating About dialog icon with Inkscape..."
	inkscape "$SVG_FILE" --export-type=png --export-filename="$PNG_FILE" -w 96 -h 96
	echo "Created: $PNG_FILE (96x96)"
else
	echo "Note: PNG icon creation skipped (install rsvg-convert, imagemagick, or inkscape for automatic conversion)"
	echo "The SVG icon will be used directly, which may be too large."
	exit 0
fi

chmod 644 "$PNG_FILE"
