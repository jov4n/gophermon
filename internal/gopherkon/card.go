package gopherkon

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

// GenerateStarterCard creates a card image with all 3 starter gophers side by side
func (g *Generator) GenerateStarterCard(gopherPaths []string, outputPath string) error {
	if len(gopherPaths) != 3 {
		return fmt.Errorf("need exactly 3 gopher paths for starter card")
	}

	// Load all 3 gopher images
	gopherImages := make([]image.Image, 3)
	maxWidth, maxHeight := 0, 0
	
	for i, path := range gopherPaths {
		img, err := g.loadImage(path)
		if err != nil {
			return fmt.Errorf("failed to load gopher %d: %w", i+1, err)
		}
		gopherImages[i] = img
		
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	// Create card dimensions
	// Card: padding + 3 gophers with spacing
	padding := 40
	spacing := 30
	cardWidth := padding*2 + maxWidth*3 + spacing*2
	cardHeight := padding*2 + maxHeight + 60 // Extra space for labels

	// Create card image with white/light background
	card := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	
	// Fill with light background
	lightGray := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	draw.Draw(card, card.Bounds(), &image.Uniform{lightGray}, image.Point{}, draw.Src)

	// Draw each gopher centered in its section
	gopherX := padding
	for _, gopherImg := range gopherImages {
		// Center the gopher horizontally in its section
		gopherBounds := gopherImg.Bounds()
		offsetX := gopherX + (maxWidth-gopherBounds.Dx())/2
		offsetY := padding + 30 // Leave space at top for label
		
		// Draw gopher
		draw.Draw(card, image.Rect(offsetX, offsetY, offsetX+gopherBounds.Dx(), offsetY+gopherBounds.Dy()),
			gopherImg, image.Point{}, draw.Over)
		
		gopherX += maxWidth + spacing
	}

	// Save the card
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, card)
}

// GenerateBattleCard creates a battle card image with enemy gopher on top and player gopher on bottom
func (g *Generator) GenerateBattleCard(enemyPath, playerPath, outputPath string) error {
	// Load enemy gopher (top)
	enemyImg, err := g.loadImage(enemyPath)
	if err != nil {
		return fmt.Errorf("failed to load enemy gopher: %w", err)
	}

	// Load player gopher (bottom)
	playerImg, err := g.loadImage(playerPath)
	if err != nil {
		return fmt.Errorf("failed to load player gopher: %w", err)
	}

	enemyBounds := enemyImg.Bounds()
	playerBounds := playerImg.Bounds()

	// Card dimensions: vertical layout with padding
	padding := 40
	spacing := 30
	cardWidth := padding*2 + max(enemyBounds.Dx(), playerBounds.Dx())
	cardHeight := padding*3 + enemyBounds.Dy() + spacing + playerBounds.Dy()

	// Create card image with light background
	card := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	lightGray := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	draw.Draw(card, card.Bounds(), &image.Uniform{lightGray}, image.Point{}, draw.Src)

	// Draw enemy gopher centered at top
	enemyX := (cardWidth - enemyBounds.Dx()) / 2
	enemyY := padding
	draw.Draw(card, image.Rect(enemyX, enemyY, enemyX+enemyBounds.Dx(), enemyY+enemyBounds.Dy()),
		enemyImg, image.Point{}, draw.Over)

	// Draw player gopher centered at bottom
	playerX := (cardWidth - playerBounds.Dx()) / 2
	playerY := padding + enemyBounds.Dy() + spacing
	draw.Draw(card, image.Rect(playerX, playerY, playerX+playerBounds.Dx(), playerY+playerBounds.Dy()),
		playerImg, image.Point{}, draw.Over)

	// Save the card
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, card)
}

// GenerateGopherCard creates a card image with N gophers arranged in a grid
func (g *Generator) GenerateGopherCard(gopherPaths []string, outputPath string, cols int) error {
	if len(gopherPaths) == 0 {
		return fmt.Errorf("need at least 1 gopher path")
	}

	// Load all gopher images
	gopherImages := make([]image.Image, len(gopherPaths))
	maxWidth, maxHeight := 0, 0
	
	for i, path := range gopherPaths {
		img, err := g.loadImage(path)
		if err != nil {
			return fmt.Errorf("failed to load gopher %d: %w", i+1, err)
		}
		gopherImages[i] = img
		
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	// Calculate grid dimensions
	numGophers := len(gopherImages)
	rows := (numGophers + cols - 1) / cols // Ceiling division

	// Create card dimensions
	padding := 30
	spacing := 20
	cardWidth := padding*2 + maxWidth*cols + spacing*(cols-1)
	cardHeight := padding*2 + maxHeight*rows + spacing*(rows-1)

	// Create card image with light background
	card := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	lightGray := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	draw.Draw(card, card.Bounds(), &image.Uniform{lightGray}, image.Point{}, draw.Src)

	// Draw gophers in grid
	for i, gopherImg := range gopherImages {
		row := i / cols
		col := i % cols
		
		gopherBounds := gopherImg.Bounds()
		offsetX := padding + col*(maxWidth+spacing) + (maxWidth-gopherBounds.Dx())/2
		offsetY := padding + row*(maxHeight+spacing) + (maxHeight-gopherBounds.Dy())/2
		
		draw.Draw(card, image.Rect(offsetX, offsetY, offsetX+gopherBounds.Dx(), offsetY+gopherBounds.Dy()),
			gopherImg, image.Point{}, draw.Over)
	}

	// Save the card
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, card)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

