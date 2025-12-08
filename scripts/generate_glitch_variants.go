package main

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gophermon-bot/internal/gopherkon"
)

func main() {
	assetsPath := "assets/artwork"
	outputDir := "glitch_variants"
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}
	
	fmt.Println("=== Generating Glitch Gopher Variants ===")
	fmt.Printf("Output directory: %s\n\n", outputDir)
	
	// Initialize generator
	generator, err := gopherkon.NewGenerator(assetsPath)
	if err != nil {
		fmt.Printf("Error loading generator: %v\n", err)
		return
	}
	
	// Set random seed
	rand.Seed(time.Now().UnixNano())
	
	// Generate 2 of each rarity (10 total)
	rarities := []string{"COMMON", "UNCOMMON", "RARE", "EPIC", "LEGENDARY"}
	
	for rarityIdx, rarity := range rarities {
		for variant := 1; variant <= 2; variant++ {
			seed := time.Now().UnixNano() + int64(rarityIdx*1000) + int64(variant*100)
			rand.Seed(seed)
			
			// Generate variant based on rarity
			layers, err := generateGlitchVariant(generator, assetsPath, rarity, seed)
			if err != nil {
				fmt.Printf("Error generating %s variant %d: %v\n", rarity, variant, err)
				continue
			}
			
			// Composite layers
			composite, err := compositeLayers(generator, layers)
			if err != nil {
				fmt.Printf("Error compositing %s variant %d: %v\n", rarity, variant, err)
				continue
			}
			
			// Save
			filename := fmt.Sprintf("glitch_%s_%d.png", strings.ToLower(rarity), variant)
			outputPath := filepath.Join(outputDir, filename)
			
			file, err := os.Create(outputPath)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				continue
			}
			
			err = png.Encode(file, composite)
			file.Close()
			if err != nil {
				fmt.Printf("Error encoding PNG: %v\n", err)
				continue
			}
			
			fmt.Printf("✓ Generated: %s (%d layers)\n", filename, len(layers))
		}
	}
	
	fmt.Printf("\n=== Complete ===\n")
	fmt.Printf("Generated 10 glitch gopher variants in: %s\n", outputDir)
}

func generateGlitchVariant(generator *gopherkon.Generator, assetsPath string, rarity string, seed int64) ([]string, error) {
	rng := rand.New(rand.NewSource(seed))
	layers := []string{}
	
	// All variants use glitch_gopher as body
	bodyPath := filepath.Join(assetsPath, "010-Body", "glitch_gopher.png")
	layers = append(layers, bodyPath)
	
	// Get available assets
	eyesPath := filepath.Join(assetsPath, "020-Eyes")
	shirtsPath := filepath.Join(assetsPath, "021-Shirts")
	hairPath := filepath.Join(assetsPath, "022-Hair")
	facialHairPath := filepath.Join(assetsPath, "023-Facial_Hair")
	glassesPath := filepath.Join(assetsPath, "024-Glasses")
	hatsPath := filepath.Join(assetsPath, "025-Hats_and_Hair_Accessories")
	extrasPath := filepath.Join(assetsPath, "027-Extras")
	
	// Helper to get random file from directory
	getRandomFile := func(dir string) (string, error) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return "", err
		}
		files := []string{}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
		if len(files) == 0 {
			return "", fmt.Errorf("no files in %s", dir)
		}
		return files[rng.Intn(len(files))], nil
	}
	
	// All get eyes + 1 extra (minimum)
	eyes, err := getRandomFile(eyesPath)
	if err == nil {
		layers = append(layers, eyes)
	}
	
	extra, err := getRandomFile(extrasPath)
	if err == nil {
		layers = append(layers, extra)
	}
	
	// Apply rarity-based rules
	switch rarity {
	case "UNCOMMON":
		// + Facial Hair OR Glasses
		if rng.Float32() < 0.5 {
			facial, err := getRandomFile(facialHairPath)
			if err == nil {
				layers = append(layers, facial)
			}
		} else {
			glass, err := getRandomFile(glassesPath)
			if err == nil {
				layers = append(layers, glass)
			}
		}
		
	case "RARE":
		// + (Hair OR Facial Hair) AND (Another Extra OR Glasses)
		if rng.Float32() < 0.5 {
			hair, err := getRandomFile(hairPath)
			if err == nil {
				layers = append(layers, hair)
			}
		} else {
			facial, err := getRandomFile(facialHairPath)
			if err == nil {
				layers = append(layers, facial)
			}
		}
		
		if rng.Float32() < 0.5 {
			extra2, err := getRandomFile(extrasPath)
			if err == nil {
				layers = append(layers, extra2)
			}
		} else {
			glass, err := getRandomFile(glassesPath)
			if err == nil {
				layers = append(layers, glass)
			}
		}
		
	case "EPIC":
		// + Hair OR Facial Hair + Glasses + Another Extra
		if rng.Float32() < 0.5 {
			hair, err := getRandomFile(hairPath)
			if err == nil {
				layers = append(layers, hair)
			}
		} else {
			facial, err := getRandomFile(facialHairPath)
			if err == nil {
				layers = append(layers, facial)
			}
		}
		
		glass, err := getRandomFile(glassesPath)
		if err == nil {
			layers = append(layers, glass)
		}
		
		extra2, err := getRandomFile(extrasPath)
		if err == nil {
			layers = append(layers, extra2)
		}
		
	case "LEGENDARY":
		// All categories + 3 Extras
		// Add shirt
		shirt, err := getRandomFile(shirtsPath)
		if err == nil {
			layers = append(layers, shirt)
		}
		
		// Add hair
		hair, err := getRandomFile(hairPath)
		if err == nil {
			layers = append(layers, hair)
		}
		
		// Add facial hair
		facial, err := getRandomFile(facialHairPath)
		if err == nil {
			layers = append(layers, facial)
		}
		
		// Add glasses
		glass, err := getRandomFile(glassesPath)
		if err == nil {
			layers = append(layers, glass)
		}
		
		// Add hat
		hat, err := getRandomFile(hatsPath)
		if err == nil {
			layers = append(layers, hat)
		}
		
		// Add 2 more extras (already have 1)
		for i := 0; i < 2; i++ {
			extra2, err := getRandomFile(extrasPath)
			if err == nil {
				layers = append(layers, extra2)
			}
		}
	}
	
	return layers, nil
}

func compositeLayers(generator *gopherkon.Generator, layerPaths []string) (image.Image, error) {
	if len(layerPaths) == 0 {
		return nil, fmt.Errorf("no layers to composite")
	}
	
	// Load first layer (body)
	baseImg, err := generator.LoadImageFromPath(layerPaths[0])
	if err != nil {
		return nil, fmt.Errorf("failed to load base image: %w", err)
	}
	
	// Composite remaining layers
	for i := 1; i < len(layerPaths); i++ {
		img, err := generator.LoadImageFromPath(layerPaths[i])
		if err != nil {
			continue // Skip layers that fail to load
		}
		
		// Resize to match base if needed
		baseBounds := baseImg.Bounds()
		imgBounds := img.Bounds()
		if imgBounds.Dx() != baseBounds.Dx() || imgBounds.Dy() != baseBounds.Dy() {
			img = resizeImage(img, baseBounds.Dx(), baseBounds.Dy())
		}
		
		baseImg = compositeLayer(baseImg, img)
	}
	
	return baseImg, nil
}

func compositeLayer(base, overlay image.Image) image.Image {
	bounds := base.Bounds()
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, base, image.Point{}, draw.Src)
	draw.Draw(result, bounds, overlay, image.Point{}, draw.Over)
	return result
}

func resizeImage(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	resized := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			srcX := x * srcWidth / targetWidth
			srcY := y * srcHeight / targetHeight
			
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

