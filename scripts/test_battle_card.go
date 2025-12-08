package main

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

	"gophermon-bot/internal/gopherkon"
)

func main() {
	// Configuration - adjust these values to align elements
	config := struct {
		// Gopher sizes
		MaxGopherSize int // Maximum width/height for gophers (they maintain aspect ratio)

		// Player gopher position (left platform)
		// These are the CENTER coordinates where the gopher will be placed
		PlayerPlatformX int // X coordinate of platform center (left side)
		PlayerPlatformY int // Y coordinate of platform center (will be calculated if 0)

		// Enemy gopher position (right platform)
		// These are the CENTER coordinates where the gopher will be placed
		EnemyPlatformX int // X coordinate of platform center (will be calculated if 0)
		EnemyPlatformY int // Y coordinate of platform center (will be calculated if 0)

		// Text positions (name plates)
		// These are the TOP-LEFT coordinates where text starts
		PlayerTextX int // X position of player name/level text (left frame)
		PlayerTextY int // Y position of player name/level text (left frame)
		EnemyTextX  int // X position of enemy name/level text (will be calculated if 0)
		EnemyTextY  int // Y position of enemy name/level text (right frame)

		// Font settings
		FontSize   int // Font size in points (e.g., 24, 32, 48, 64)
		LineHeight int // Space between text lines in pixels (auto-calculated if 0)

		// Test data
		PlayerName  string
		PlayerLevel int
		EnemyName   string
		EnemyLevel  int
	}{
		MaxGopherSize:   300,
		PlayerPlatformX: 360,
		PlayerPlatformY: 720,  // Will be calculated as cardHeight - 200 if 0
		EnemyPlatformX:  1120, // Will be calculated as cardWidth - 300 if 0
		EnemyPlatformY:  720,  // Will be calculated as cardHeight - 200 if 0
		PlayerTextX:     150,
		PlayerTextY:     130,
		EnemyTextX:      935, // Will be calculated as cardWidth - 420 if 0
		EnemyTextY:      0,
		FontSize:        32, // Font size in points (e.g., 24, 32, 48, 64)
		LineHeight:      0,  // Auto-calculated based on font size if 0
		PlayerName:      "HackerGopher",
		PlayerLevel:     1,
		EnemyName:       "Bytebit",
		EnemyLevel:      6,
	}

	// Initialize generator
	assetsPath := "assets/artwork"
	generator, err := gopherkon.NewGenerator(assetsPath)
	if err != nil {
		fmt.Printf("Error loading generator: %v\n", err)
		return
	}

	// Generate real gopher sprites
	fmt.Println("Generating player gopher...")
	playerResult, err := generator.Generate(gopherkon.GenerateOptions{
		Complexity:   3,
		TargetRarity: "UNCOMMON",
		Seed:         12345, // Fixed seed for consistent testing
	})
	if err != nil {
		fmt.Printf("Error generating player gopher: %v\n", err)
		return
	}
	playerImg := playerResult.Image

	fmt.Println("Generating enemy gopher...")
	enemyResult, err := generator.Generate(gopherkon.GenerateOptions{
		Complexity:   2,
		TargetRarity: "COMMON",
		Seed:         54321, // Different seed for variety
	})
	if err != nil {
		fmt.Printf("Error generating enemy gopher: %v\n", err)
		return
	}
	enemyImg := enemyResult.Image

	// Load battle screen
	assetsDir := filepath.Dir(assetsPath)
	battleScreenPath := filepath.Join(assetsDir, "battle_screen.png")
	battleScreen, err := loadImage(battleScreenPath)
	if err != nil {
		fmt.Printf("Error loading battle screen: %v\n", err)
		return
	}

	battleBounds := battleScreen.Bounds()
	cardWidth := battleBounds.Dx()
	cardHeight := battleBounds.Dy()

	// Update calculated positions (only if set to 0)
	if config.PlayerPlatformY == 0 {
		config.PlayerPlatformY = cardHeight - 200
	}
	if config.EnemyPlatformX == 0 {
		config.EnemyPlatformX = cardWidth - 300
	}
	if config.EnemyPlatformY == 0 {
		config.EnemyPlatformY = cardHeight - 200
	}
	if config.EnemyTextX == 0 {
		config.EnemyTextX = cardWidth - 420
	}
	if config.EnemyTextY == 0 {
		config.EnemyTextY = config.PlayerTextY // Match player text Y by default
	}

	fmt.Printf("\nBattle screen dimensions: %dx%d\n", cardWidth, cardHeight)
	fmt.Printf("Player platform: (%d, %d)\n", config.PlayerPlatformX, config.PlayerPlatformY)
	fmt.Printf("Enemy platform: (%d, %d)\n", config.EnemyPlatformX, config.EnemyPlatformY)
	fmt.Printf("Player text: (%d, %d)\n", config.PlayerTextX, config.PlayerTextY)
	fmt.Printf("Enemy text: (%d, %d)\n", config.EnemyTextX, config.EnemyTextY)
	fmt.Printf("Player gopher size: %dx%d\n", playerImg.Bounds().Dx(), playerImg.Bounds().Dy())
	fmt.Printf("Enemy gopher size: %dx%d\n", enemyImg.Bounds().Dx(), enemyImg.Bounds().Dy())

	// Create card
	card := image.NewRGBA(image.Rect(0, 0, cardWidth, cardHeight))
	draw.Draw(card, card.Bounds(), battleScreen, image.Point{}, draw.Src)

	// Resize gophers
	playerScaled := resizeImage(playerImg, config.MaxGopherSize, config.MaxGopherSize)
	playerScaledBounds := playerScaled.Bounds()

	enemyScaled := resizeImage(enemyImg, config.MaxGopherSize, config.MaxGopherSize)
	enemyScaledBounds := enemyScaled.Bounds()

	// Draw player gopher (centered on platform)
	playerX := config.PlayerPlatformX - playerScaledBounds.Dx()/2
	playerY := config.PlayerPlatformY - playerScaledBounds.Dy()/2
	draw.Draw(card, image.Rect(playerX, playerY, playerX+playerScaledBounds.Dx(), playerY+playerScaledBounds.Dy()),
		playerScaled, image.Point{}, draw.Over)

	// Draw enemy gopher (centered on platform)
	enemyX := config.EnemyPlatformX - enemyScaledBounds.Dx()/2
	enemyY := config.EnemyPlatformY - enemyScaledBounds.Dy()/2
	draw.Draw(card, image.Rect(enemyX, enemyY, enemyX+enemyScaledBounds.Dx(), enemyY+enemyScaledBounds.Dy()),
		enemyScaled, image.Point{}, draw.Over)

	// Draw text
	cyanColor := color.RGBA{R: 0, G: 255, B: 255, A: 255}
	playerText := fmt.Sprintf("%s\nLv.%d", config.PlayerName, config.PlayerLevel)
	enemyText := fmt.Sprintf("%s\nLv.%d", config.EnemyName, config.EnemyLevel)

	// Calculate line height if not set
	lineHeight := config.LineHeight
	if lineHeight == 0 {
		lineHeight = config.FontSize + 8 // Auto-calculate based on font size
	}

	drawTextMultiline(card, playerText, config.PlayerTextX, config.PlayerTextY, cyanColor, config.FontSize, lineHeight)
	drawTextMultiline(card, enemyText, config.EnemyTextX, config.EnemyTextY, cyanColor, config.FontSize, lineHeight)

	// Save result
	outputPath := "test_battle_card.png"
	file, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, card); err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	fmt.Printf("\nTest battle card saved to: %s\n", outputPath)
	fmt.Println("\nTo adjust positions, edit the config struct in this script and run again.")
	fmt.Println("Run with: go run scripts/test_battle_card.go")
}

func loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	return img, err
}

func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	scaleX := float64(maxWidth) / float64(width)
	scaleY := float64(maxHeight) / float64(height)
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) / scale)
			srcY := int(float64(y) / scale)
			if srcX < width && srcY < height {
				resized.Set(x, y, img.At(srcX, srcY))
			}
		}
	}

	return resized
}

func drawTextMultiline(img *image.RGBA, text string, x, y int, clr color.Color, fontSize int, lineHeight int) {
	lines := strings.Split(text, "\n")
	currentY := y

	// Try to auto-detect a system font
	fontPath := findSystemFont()

	if fontPath != "" {
		// Use freetype for scalable TrueType fonts
		for _, line := range lines {
			if line != "" {
				err := drawTextWithFreetype(img, line, x, currentY, clr, fontSize, fontPath)
				if err != nil {
					fmt.Printf("Warning: Could not draw text with freetype: %v. Using scaled basicfont.\n", err)
					// Fall back to scaled basicfont
					drawTextScaled(img, line, x, currentY, clr, fontSize)
				}
				currentY += lineHeight
			}
		}
	} else {
		// Use scaled basicfont
		for _, line := range lines {
			if line != "" {
				drawTextScaled(img, line, x, currentY, clr, fontSize)
			}
			currentY += lineHeight
		}
	}
}

// findSystemFont tries to find a system font automatically
func findSystemFont() string {
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
func drawTextScaled(img *image.RGBA, text string, x, y int, clr color.Color, fontSize int) {
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
	scaledImg := resizeImage(tempImg, scaledWidth, scaledHeight)

	// Draw scaled text onto main image
	draw.Draw(img, image.Rect(x, y, x+scaledWidth, y+scaledHeight),
		scaledImg, image.Point{}, draw.Over)
}

func drawTextBasic(img *image.RGBA, text string, x, y int, clr color.Color) {
	point := fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6((y + 13) * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(clr),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(text)
}

// drawTextWithFreetype draws text using freetype for scalable fonts
func drawTextWithFreetype(img *image.RGBA, text string, x, y int, clr color.Color, fontSize int, fontPath string) error {
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
