package gopherkon

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	ft "github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
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

// GenerateStarterCardFromImages creates a card from image.Image objects (for base64 support)
func (g *Generator) GenerateStarterCardFromImages(gopherImages []image.Image, outputPath string) error {
	card, err := g.generateStarterCardImage(gopherImages)
	if err != nil {
		return err
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

// GenerateStarterCardFromImagesToBase64 creates a card from image.Image objects and returns base64
func (g *Generator) GenerateStarterCardFromImagesToBase64(gopherImages []image.Image) (string, error) {
	card, err := g.generateStarterCardImage(gopherImages)
	if err != nil {
		return "", err
	}
	return g.EncodeImageToBase64(card)
}

// generateStarterCardImage creates the card image (internal helper)
func (g *Generator) generateStarterCardImage(gopherImages []image.Image) (image.Image, error) {
	if len(gopherImages) != 3 {
		return nil, fmt.Errorf("need exactly 3 gopher images for starter card")
	}

	maxWidth, maxHeight := 0, 0
	for _, img := range gopherImages {
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
	}

	// Create card dimensions
	padding := 40
	spacing := 30
	cardWidth := padding*2 + maxWidth*3 + spacing*2
	cardHeight := padding*2 + maxHeight + 60

	// Create card image with white/light background
	card := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	lightGray := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	draw.Draw(card, card.Bounds(), &image.Uniform{lightGray}, image.Point{}, draw.Src)

	// Draw each gopher centered in its section
	gopherX := padding
	for _, gopherImg := range gopherImages {
		gopherBounds := gopherImg.Bounds()
		offsetX := gopherX + (maxWidth-gopherBounds.Dx())/2
		offsetY := padding + 30
		
		draw.Draw(card, image.Rect(offsetX, offsetY, offsetX+gopherBounds.Dx(), offsetY+gopherBounds.Dy()),
			gopherImg, image.Point{}, draw.Over)
		
		gopherX += maxWidth + spacing
	}

	return card, nil
}

// GenerateBattleCardFromImages creates a battle card from image.Image objects (for base64 support)
func (g *Generator) GenerateBattleCardFromImages(enemyImg, playerImg image.Image, enemyName, playerName string, enemyLevel, playerLevel int, outputPath string) error {
	card, err := g.generateBattleCardImage(enemyImg, playerImg, enemyName, playerName, enemyLevel, playerLevel)
	if err != nil {
		return err
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

// GenerateBattleCardFromImagesToBase64 creates a battle card from image.Image objects and returns base64
func (g *Generator) GenerateBattleCardFromImagesToBase64(enemyImg, playerImg image.Image, enemyName, playerName string, enemyLevel, playerLevel int) (string, error) {
	card, err := g.generateBattleCardImage(enemyImg, playerImg, enemyName, playerName, enemyLevel, playerLevel)
	if err != nil {
		return "", err
	}
	return g.EncodeImageToBase64(card)
}

// generateBattleCardImage creates the battle card image using the battle screen background
func (g *Generator) generateBattleCardImage(enemyImg, playerImg image.Image, enemyName, playerName string, enemyLevel, playerLevel int) (image.Image, error) {
	// Load battle screen background
	// assetsPath is "assets/artwork", so go up one level to get "assets"
	assetsDir := filepath.Dir(g.assetsPath)
	battleScreenPath := filepath.Join(assetsDir, "battle_screen.png")
	battleScreen, err := g.loadImage(battleScreenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load battle screen from %s: %w", battleScreenPath, err)
	}

	battleBounds := battleScreen.Bounds()
	cardWidth := battleBounds.Dx()
	cardHeight := battleBounds.Dy()

	// Create card image and draw battle screen background
	card := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	draw.Draw(card, card.Bounds(), battleScreen, image.Point{}, draw.Src)

	// Resize gophers to fit on circular platforms
	maxGopherSize := 300

	// Resize enemy gopher
	enemyScaled := g.resizeImage(enemyImg, maxGopherSize, maxGopherSize)
	enemyScaledBounds := enemyScaled.Bounds()

	// Resize player gopher
	playerScaled := g.resizeImage(playerImg, maxGopherSize, maxGopherSize)
	playerScaledBounds := playerScaled.Bounds()

	// Position gophers on circular platforms - using values from test script
	// Player gopher on left platform (centered on platform)
	playerPlatformX := 360
	playerPlatformY := 720
	playerX := playerPlatformX - playerScaledBounds.Dx()/2
	playerY := playerPlatformY - playerScaledBounds.Dy()/2
	draw.Draw(card, image.Rect(playerX, playerY, playerX+playerScaledBounds.Dx(), playerY+playerScaledBounds.Dy()),
		playerScaled, image.Point{}, draw.Over)

	// Enemy gopher on right platform (centered on platform)
	enemyPlatformX := 1120
	enemyPlatformY := 720
	enemyX := enemyPlatformX - enemyScaledBounds.Dx()/2
	enemyY := enemyPlatformY - enemyScaledBounds.Dy()/2
	draw.Draw(card, image.Rect(enemyX, enemyY, enemyX+enemyScaledBounds.Dx(), enemyY+enemyScaledBounds.Dy()),
		enemyScaled, image.Point{}, draw.Over)

	// Add names and levels to rectangular frames at top - using values from test script
	cyanColor := color.RGBA{R: 0, G: 255, B: 255, A: 255}
	fontSize := 32
	lineHeight := fontSize + 8 // Auto-calculated based on font size
	
	// Player name and level (left frame)
	playerText := fmt.Sprintf("%s\nLv.%d", playerName, playerLevel)
	g.drawTextMultilineScaled(card, playerText, 150, 130, cyanColor, fontSize, lineHeight)
	
	// Enemy name and level (right frame)
	enemyText := fmt.Sprintf("%s\nLv.%d", enemyName, enemyLevel)
	g.drawTextMultilineScaled(card, enemyText, 935, 130, cyanColor, fontSize, lineHeight)

	return card, nil
}

// resizeImage resizes an image to fit within maxWidth and maxHeight while maintaining aspect ratio
func (g *Generator) resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate scaling factor to fit within max dimensions
	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	// Create resized image
	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	
	// Bilinear interpolation for better quality
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := float64(x) / scale
			srcY := float64(y) / scale
			
			// Get integer and fractional parts
			x0 := int(srcX)
			y0 := int(srcY)
			x1 := x0 + 1
			y1 := y0 + 1
			
			// Clamp to image bounds
			if x1 >= width {
				x1 = width - 1
			}
			if y1 >= height {
				y1 = height - 1
			}
			if x0 >= width {
				x0 = width - 1
			}
			if y0 >= height {
				y0 = height - 1
			}
			
			fx := srcX - float64(x0)
			fy := srcY - float64(y0)
			
			// Get four corner colors
			c00 := img.At(x0, y0)
			c10 := img.At(x1, y0)
			c01 := img.At(x0, y1)
			c11 := img.At(x1, y1)
			
			// Interpolate
			c0 := g.interpolateColor(c00, c10, fx)
			c1 := g.interpolateColor(c01, c11, fx)
			finalColor := g.interpolateColor(c0, c1, fy)
			
			resized.Set(x, y, finalColor)
		}
	}

	return resized
}

// interpolateColor interpolates between two colors
func (g *Generator) interpolateColor(c1, c2 color.Color, t float64) color.Color {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	
	r := uint8(float64(r1>>8)*(1-t) + float64(r2>>8)*t)
	green := uint8(float64(g1>>8)*(1-t) + float64(g2>>8)*t)
	b := uint8(float64(b1>>8)*(1-t) + float64(b2>>8)*t)
	a := uint8(float64(a1>>8)*(1-t) + float64(a2>>8)*t)
	
	return color.RGBA{R: r, G: green, B: b, A: a}
}

// drawText draws text on an image at the specified position
func (g *Generator) drawText(img *image.RGBA, text string, x, y int, clr color.Color) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6((y + 13) * 64)} // Adjust Y for baseline
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(clr),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)
}

// drawTextMultiline draws multiline text on an image (legacy, uses basicfont)
func (g *Generator) drawTextMultiline(img *image.RGBA, text string, x, y int, clr color.Color) {
	lines := strings.Split(text, "\n")
	lineHeight := 16 // Space between lines
	currentY := y
	for _, line := range lines {
		if line != "" {
			g.drawText(img, line, x, currentY, clr)
		}
		currentY += lineHeight
	}
}

// drawTextMultilineScaled draws multiline text with scalable fonts
func (g *Generator) drawTextMultilineScaled(img *image.RGBA, text string, x, y int, clr color.Color, fontSize int, lineHeight int) {
	lines := strings.Split(text, "\n")
	currentY := y

	// Try to auto-detect a system font
	fontPath := g.findSystemFont()

	if fontPath != "" {
		// Use freetype for scalable TrueType fonts
		for _, line := range lines {
			if line != "" {
				err := g.drawTextWithFreetype(img, line, x, currentY, clr, fontSize, fontPath)
				if err != nil {
					// Fall back to scaled basicfont
					g.drawTextScaled(img, line, x, currentY, clr, fontSize)
				}
				currentY += lineHeight
			}
		}
	} else {
		// Use scaled basicfont
		for _, line := range lines {
			if line != "" {
				g.drawTextScaled(img, line, x, currentY, clr, fontSize)
			}
			currentY += lineHeight
		}
	}
}

// findSystemFont tries to find a system font automatically
func (g *Generator) findSystemFont() string {
	// Common Windows font paths
	windowsFonts := []string{
		"C:/Windows/Fonts/arial.ttf",
		"C:/Windows/Fonts/calibri.ttf",
		"C:/Windows/Fonts/comic.ttf",
		"C:/Windows/Fonts/times.ttf",
		"C:/Windows/Fonts/cour.ttf",
	}

	for _, fontPath := range windowsFonts {
		if _, err := os.Stat(fontPath); err == nil {
			return fontPath
		}
	}

	return "" // No system font found
}

// drawTextScaled draws text using scaled basicfont (bitmap scaling)
func (g *Generator) drawTextScaled(img *image.RGBA, text string, x, y int, clr color.Color, fontSize int) {
	// Calculate scale factor (basicfont is 7x13, so scale to desired size)
	scale := float64(fontSize) / 13.0
	if scale < 1.0 {
		scale = 1.0
	}

	// Create a temporary image for the text at original size
	tempWidth := len(text) * 7
	tempHeight := 13
	tempImg := image.NewRGBA(image.Rect(0, 0, tempWidth, tempHeight))

	// Draw text at original size
	point := fixed.Point26_6{X: fixed.Int26_6(0 * 64), Y: fixed.Int26_6(13 * 64)}
	d := &font.Drawer{
		Dst:  tempImg,
		Src:  image.NewUniform(clr),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)

	// Scale the text image
	scaledWidth := int(float64(tempWidth) * scale)
	scaledHeight := int(float64(tempHeight) * scale)
	scaledImg := g.resizeImage(tempImg, scaledWidth, scaledHeight)

	// Draw scaled text onto main image
	draw.Draw(img, image.Rect(x, y, x+scaledWidth, y+scaledHeight),
		scaledImg, image.Point{}, draw.Over)
}

// drawTextWithFreetype draws text using freetype for scalable fonts
func (g *Generator) drawTextWithFreetype(img *image.RGBA, text string, x, y int, clr color.Color, fontSize int, fontPath string) error {
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return fmt.Errorf("failed to read font file: %w", err)
	}

	ttfFont, err := truetype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("failed to parse font: %w", err)
	}

	c := ft.NewContext()
	c.SetDPI(72)
	c.SetFont(ttfFont)
	c.SetFontSize(float64(fontSize))
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(clr))

	pt := ft.Pt(x, y+fontSize) // Adjust Y for baseline
	_, err = c.DrawString(text, pt)
	return err
}

// GenerateGopherCardFromImages creates a card from image.Image objects (for base64 support)
func (g *Generator) GenerateGopherCardFromImages(gopherImages []image.Image, outputPath string, cols int) error {
	card, err := g.generateGopherCardImage(gopherImages, cols)
	if err != nil {
		return err
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

// GenerateGopherCardFromImagesToBase64 creates a card from image.Image objects and returns base64
func (g *Generator) GenerateGopherCardFromImagesToBase64(gopherImages []image.Image, cols int) (string, error) {
	card, err := g.generateGopherCardImage(gopherImages, cols)
	if err != nil {
		return "", err
	}
	return g.EncodeImageToBase64(card)
}

// generateGopherCardImage creates the gopher card image (internal helper)
func (g *Generator) generateGopherCardImage(gopherImages []image.Image, cols int) (image.Image, error) {
	if len(gopherImages) == 0 {
		return nil, fmt.Errorf("need at least 1 gopher image")
	}

	maxWidth, maxHeight := 0, 0
	for _, img := range gopherImages {
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
	rows := (numGophers + cols - 1) / cols

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

	return card, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

