package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"gophermon-bot/internal/gopherkon"
)

func main() {
	// Command-line flags
	var (
		body       = flag.String("body", "", "Body asset file (e.g., glitch_gopher.png or full path)")
		eyes       = flag.String("eyes", "", "Eyes asset file")
		shirt      = flag.String("shirt", "", "Shirt asset file")
		hair       = flag.String("hair", "", "Hair asset file")
		facialHair = flag.String("facial", "", "Facial hair asset file")
		glasses    = flag.String("glasses", "", "Glasses asset file")
		hat        = flag.String("hat", "", "Hat asset file")
		extra      = flag.String("extra", "", "Extra asset file (can specify multiple with comma)")
		output     = flag.String("output", "test_custom.png", "Output file")
		assetsPath = flag.String("assets", "assets/artwork", "Assets directory path")
		list       = flag.Bool("list", false, "List available assets by category")
	)
	flag.Parse()

	// Initialize generator to get asset paths
	generator, err := gopherkon.NewGenerator(*assetsPath)
	if err != nil {
		fmt.Printf("Error loading generator: %v\n", err)
		return
	}

	// List available assets if requested
	if *list {
		listAvailableAssets(generator, *assetsPath)
		return
	}

	// Collect all specified layers
	layers := []string{}

	// Add layers in order (body first, then eyes, etc.)
	if *body != "" {
		bodyPath := resolveAssetPath(*body, *assetsPath, "010-Body", "020-Body")
		if bodyPath != "" {
			layers = append(layers, bodyPath)
		} else {
			fmt.Printf("Warning: Could not find body asset: %s (searched in 010-Body, 020-Body)\n", *body)
		}
	}

	if *eyes != "" {
		eyesPath := resolveAssetPath(*eyes, *assetsPath, "020-Eyes")
		if eyesPath != "" {
			layers = append(layers, eyesPath)
		} else {
			fmt.Printf("Warning: Could not find eyes asset: %s\n", *eyes)
		}
	}

	if *shirt != "" {
		shirtPath := resolveAssetPath(*shirt, *assetsPath, "021-Shirts")
		if shirtPath != "" {
			layers = append(layers, shirtPath)
		} else {
			fmt.Printf("Warning: Could not find shirt asset: %s\n", *shirt)
		}
	}

	if *hair != "" {
		hairPath := resolveAssetPath(*hair, *assetsPath, "022-Hair")
		if hairPath != "" {
			layers = append(layers, hairPath)
		} else {
			fmt.Printf("Warning: Could not find hair asset: %s\n", *hair)
		}
	}

	if *facialHair != "" {
		facialPath := resolveAssetPath(*facialHair, *assetsPath, "023-Facial_Hair")
		if facialPath != "" {
			layers = append(layers, facialPath)
		} else {
			fmt.Printf("Warning: Could not find facial hair asset: %s\n", *facialHair)
		}
	}

	if *glasses != "" {
		glassesPath := resolveAssetPath(*glasses, *assetsPath, "024-Glasses")
		if glassesPath != "" {
			layers = append(layers, glassesPath)
		} else {
			fmt.Printf("Warning: Could not find glasses asset: %s\n", *glasses)
		}
	}

	if *hat != "" {
		hatPath := resolveAssetPath(*hat, *assetsPath, "025-Hats_and_Hair_Accessories")
		if hatPath != "" {
			layers = append(layers, hatPath)
		} else {
			fmt.Printf("Warning: Could not find hat asset: %s\n", *hat)
		}
	}

	// Handle extras (can be multiple)
	if *extra != "" {
		extras := strings.Split(*extra, ",")
		for _, ext := range extras {
			ext = strings.TrimSpace(ext)
			if ext != "" {
				extraPath := resolveAssetPath(ext, *assetsPath, "027-Extras")
				if extraPath != "" {
					layers = append(layers, extraPath)
				} else {
					fmt.Printf("Warning: Could not find extra asset: %s\n", ext)
				}
			}
		}
	}

	if len(layers) == 0 {
		fmt.Println("No layers specified! Use -list to see available assets.")
		fmt.Println("\nExample:")
		fmt.Println("  go run scripts/test_custom_artwork.go -body=glitch_gopher.png -eyes=crazy_eyes.png")
		return
	}

	// Composite layers
	fmt.Println("=== Custom Artwork Generation ===")
	fmt.Printf("Layers to composite: %d\n\n", len(layers))

	// Load and composite images
	var baseImg image.Image
	var targetWidth, targetHeight int
	
	for i, layerPath := range layers {
		fmt.Printf("Loading layer %d: %s\n", i+1, filepath.Base(layerPath))
		img, err := generator.LoadImageFromPath(layerPath)
		if err != nil {
			fmt.Printf("Error loading %s: %v\n", layerPath, err)
			continue
		}

		if i == 0 {
			// First layer (body) - determine target size
			// Standard body size is 1300x1392 (matches other body assets)
			bounds := img.Bounds()
			currentWidth := bounds.Dx()
			currentHeight := bounds.Dy()
			
			// Resize body to standard size if needed
			targetWidth = 1300
			targetHeight = 1392
			if currentWidth != targetWidth || currentHeight != targetHeight {
				fmt.Printf("  Resizing body from %dx%d to %dx%d\n", currentWidth, currentHeight, targetWidth, targetHeight)
				img = resizeImage(img, targetWidth, targetHeight)
			}
			
			baseImg = img
		} else {
			// Resize other layers to match base image size
			bounds := img.Bounds()
			if bounds.Dx() != targetWidth || bounds.Dy() != targetHeight {
				fmt.Printf("  Resizing layer to match base (%dx%d)\n", targetWidth, targetHeight)
				img = resizeImage(img, targetWidth, targetHeight)
			}
			baseImg = compositeLayer(baseImg, img)
		}
	}

	// Save result
	if baseImg == nil {
		fmt.Println("Error: No valid layers could be loaded")
		return
	}

	outputFile := *output
	if !strings.HasSuffix(outputFile, ".png") {
		outputFile = outputFile + ".png"
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	err = png.Encode(file, baseImg)
	if err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	fmt.Printf("\n✓ Generated: %s\n", outputFile)
	fmt.Printf("  Dimensions: %dx%d\n", baseImg.Bounds().Dx(), baseImg.Bounds().Dy())
}

// resolveAssetPath tries to find an asset file
// First checks if it's a full path, then searches in specified category folders
func resolveAssetPath(filename string, assetsPath string, categoryFolders ...string) string {
	// Clean filename (remove path separators if present)
	cleanFilename := filepath.Base(filename)
	
	// If it's already a full path and exists, use it
	if filepath.IsAbs(filename) || (strings.Contains(filename, string(filepath.Separator)) && !strings.Contains(filename, assetsPath)) {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	// Search in specified category folders first
	for _, category := range categoryFolders {
		fullPath := filepath.Join(assetsPath, category, cleanFilename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	// Search in all subdirectories
	entries, err := os.ReadDir(assetsPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				testPath := filepath.Join(assetsPath, entry.Name(), cleanFilename)
				if _, err := os.Stat(testPath); err == nil {
					return testPath
				}
			}
		}
	}

	// Try case-insensitive search
	entries, err = os.ReadDir(assetsPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				catPath := filepath.Join(assetsPath, entry.Name())
				catEntries, err := os.ReadDir(catPath)
				if err == nil {
					for _, fileEntry := range catEntries {
						if !fileEntry.IsDir() && strings.EqualFold(fileEntry.Name(), cleanFilename) {
							return filepath.Join(catPath, fileEntry.Name())
						}
					}
				}
			}
		}
	}

	return ""
}

// compositeLayer overlays one image on top of another
func compositeLayer(base, overlay image.Image) image.Image {
	bounds := base.Bounds()
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, base, image.Point{}, draw.Src)
	draw.Draw(result, bounds, overlay, image.Point{}, draw.Over)
	return result
}

// resizeImage resizes an image to the target dimensions
func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Use nearest-neighbor for simplicity (can be upgraded to bilinear if needed)
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := x * srcWidth / targetWidth
			srcY := y * srcHeight / targetHeight
			
			// Clamp to source bounds
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}
			
			resized.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return resized
}

// listAvailableAssets lists all available assets by category
func listAvailableAssets(generator *gopherkon.Generator, assetsPath string) {
	fmt.Println("=== Available Assets ===")
	fmt.Println()

	// Get categories from generator (we'll need to access them)
	// For now, manually list common categories
	categories := []struct {
		name string
		path string
	}{
		{"Body", "010-Body"},
		{"Eyes", "020-Eyes"},
		{"Shirts", "021-Shirts"},
		{"Hair", "022-Hair"},
		{"Facial Hair", "023-Facial_Hair"},
		{"Glasses", "024-Glasses"},
		{"Hats", "025-Hats_and_Hair_Accessories"},
		{"Extras", "027-Extras"},
	}

	for _, cat := range categories {
		catPath := filepath.Join(assetsPath, cat.path)
		entries, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}

		files := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
				files = append(files, entry.Name())
			}
		}

		if len(files) > 0 {
			fmt.Printf("%s (%d files):\n", cat.name, len(files))
			// Show first 10 files
			maxShow := 10
			if len(files) < maxShow {
				maxShow = len(files)
			}
			for i := 0; i < maxShow; i++ {
				fmt.Printf("  - %s\n", files[i])
			}
			if len(files) > maxShow {
				fmt.Printf("  ... and %d more\n", len(files)-maxShow)
			}
			fmt.Println()
		}
	}
}

