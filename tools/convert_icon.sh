#!/bin/bash
# Convert SVG to PNG using ImageMagick or rsvg-convert
# Usage: ./convert_icon.sh input.svg output.png size

SVG_FILE=$1
PNG_FILE=$2
SIZE=${3:-512}

if command -v rsvg-convert &> /dev/null; then
    rsvg-convert -w $SIZE -h $SIZE "$SVG_FILE" -o "$PNG_FILE"
elif command -v inkscape &> /dev/null; then
    inkscape "$SVG_FILE" --export-type=png --export-filename="$PNG_FILE" -w $SIZE -h $SIZE
elif command -v convert &> /dev/null; then
    convert -background none -resize ${SIZE}x${SIZE} "$SVG_FILE" "$PNG_FILE"
else
    echo "Error: No SVG converter found. Please install rsvg-convert, inkscape, or imagemagick"
    exit 1
fi

echo "Created $PNG_FILE at ${SIZE}x${SIZE}px"
