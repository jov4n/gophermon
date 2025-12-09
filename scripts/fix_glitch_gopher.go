package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"gophermon-bot/internal/gopherkon"
)

func main() {
	assetsPath := "assets/artwork"
	bodyPath := filepath.Join(assetsPath, "010-Body", "pixel_gopher.png")

	fmt.Println("=== Fixing pixel_gopher.png ===")
	fmt.Printf("Loading: %s\n", bodyPath)

	// Initialize generator to use its image loading
	generator, err := gopherkon.NewGenerator(assetsPath)
	if err != nil {
		fmt.Printf("Error loading generator: %v\n", err)
		return
	}

	// Load the image
	img, err := generator.LoadImageFromPath(bodyPath)
	if err != nil {
		fmt.Printf("Error loading image: %v\n", err)
		return
	}

	bounds := img.Bounds()
	currentWidth := bounds.Dx()
	currentHeight := bounds.Dy()

	fmt.Printf("Current size: %dx%d\n", currentWidth, currentHeight)

	// Target size
	targetWidth := 1300
	targetHeight := 1392

	if currentWidth == targetWidth && currentHeight == targetHeight {
		fmt.Println("Image is already the correct size!")
		return
	}

	fmt.Printf("Resizing to: %dx%d\n", targetWidth, targetHeight)

	// Resize the image
	resized := resizeImage(img, targetWidth, targetHeight)

	// Create backup
	backupPath := bodyPath + ".backup"
	fmt.Printf("Creating backup: %s\n", backupPath)

	// Copy original to backup
	originalFile, err := os.Open(bodyPath)
	if err != nil {
		fmt.Printf("Warning: Could not create backup: %v\n", err)
	} else {
		backupFile, err := os.Create(backupPath)
		if err == nil {
			// Simple copy
			buf := make([]byte, 1024*1024) // 1MB buffer
			for {
				n, err := originalFile.Read(buf)
				if n == 0 || err != nil {
					break
				}
				backupFile.Write(buf[:n])
			}
			backupFile.Close()
			fmt.Println("Backup created successfully")
		}
		originalFile.Close()
	}

	// Save resized image
	fmt.Printf("Saving resized image: %s\n", bodyPath)
	file, err := os.Create(bodyPath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, resized)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	fmt.Println("✓ Successfully resized and saved pixel_gopher.png")
	fmt.Printf("  Original: %dx%d\n", currentWidth, currentHeight)
	fmt.Printf("  New size: %dx%d\n", targetWidth, targetHeight)
	fmt.Printf("  Backup saved to: %s\n", backupPath)
}

// resizeImage resizes an image to the target dimensions using bilinear interpolation
func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Bilinear interpolation for better quality
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			// Map target coordinates to source coordinates
			srcX := float64(x) * float64(srcWidth) / float64(targetWidth)
			srcY := float64(y) * float64(srcHeight) / float64(targetHeight)

			// Get integer and fractional parts
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1

			// Clamp to source bounds
			if x1 >= srcWidth {
				x1 = srcWidth - 1
			}
			if y1 >= srcHeight {
				y1 = srcHeight - 1
			}
			if x0 >= srcWidth {
				x0 = srcWidth - 1
			}
			if y0 >= srcHeight {
				y0 = srcHeight - 1
			}

			fx := srcX - float64(x0)
			fy := srcY - float64(y0)

			// Get four corner colors
			c00 := img.At(bounds.Min.X+x0, bounds.Min.Y+y0)
			c10 := img.At(bounds.Min.X+x1, bounds.Min.Y+y0)
			c01 := img.At(bounds.Min.X+x0, bounds.Min.Y+y1)
			c11 := img.At(bounds.Min.X+x1, bounds.Min.Y+y1)

			// Interpolate
			c0 := interpolateColor(c00, c10, fx)
			c1 := interpolateColor(c01, c11, fx)
			finalColor := interpolateColor(c0, c1, fy)

			resized.Set(x, y, finalColor)
		}
	}

	return resized
}

// interpolateColor interpolates between two colors
func interpolateColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	r := uint8(float64(r1>>8)*(1-t) + float64(r2>>8)*t)
	green := uint8(float64(g1>>8)*(1-t) + float64(g2>>8)*t)
	b := uint8(float64(b1>>8)*(1-t) + float64(b2>>8)*t)
	a := uint8(float64(a1>>8)*(1-t) + float64(a2>>8)*t)

	return color.RGBA{R: r, G: green, B: b, A: a}
}
