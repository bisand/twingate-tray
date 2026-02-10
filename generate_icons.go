//go:build ignore

package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
)

func main() {
	// Generate connected icon (green)
	generateIcon("assets/connected.png", color.RGBA{0, 200, 0, 255})

	// Generate disconnected icon (red)
	generateIcon("assets/disconnected.png", color.RGBA{200, 0, 0, 255})

	log.Println("Icons generated successfully")
}

func generateIcon(filename string, fillColor color.Color) {
	// Create a 24x24 image
	img := image.NewRGBA(image.Rect(0, 0, 24, 24))

	// Fill background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{240, 240, 240, 255}}, image.Point{}, draw.Src)

	// Draw a circle in the middle
	for x := 0; x < 24; x++ {
		for y := 0; y < 24; y++ {
			dx := float64(x - 12)
			dy := float64(y - 12)
			dist := dx*dx + dy*dy
			if dist < 100 { // radius ~10px
				img.Set(x, y, fillColor)
			}
		}
	}

	// Encode to PNG
	buf := &bytes.Buffer{}
	png.Encode(buf, img)

	// Write to file
	err := os.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Failed to write %s: %v", filename, err)
	}
}
