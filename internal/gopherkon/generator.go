package gopherkon

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// CategoryType represents the type of category
type CategoryType string

const (
	CategoryTypeBody       CategoryType = "body"
	CategoryTypeEyes       CategoryType = "eyes"
	CategoryTypeShirts     CategoryType = "shirts"
	CategoryTypeHair       CategoryType = "hair"
	CategoryTypeFacialHair CategoryType = "facial_hair"
	CategoryTypeGlasses    CategoryType = "glasses"
	CategoryTypeHats       CategoryType = "hats"
	CategoryTypeExtras     CategoryType = "extras"
	CategoryTypeUnknown    CategoryType = "unknown"
)

// CategoryInfo represents a gopherize.me category with its order and features
type CategoryInfo struct {
	Order    int          // Order number from folder name (e.g., 000, 010, 020)
	Name     string       // Category name without number prefix
	Type     CategoryType // Type of category
	Features []string     // List of PNG file paths in this category
}

// Generator handles sprite generation from gopherize.me assets
type Generator struct {
	assetsPath string
	categories []CategoryInfo // Categories sorted by order
}

// NewGenerator creates a new sprite generator for gopherize.me artwork
func NewGenerator(assetsPath string) (*Generator, error) {
	gen := &Generator{
		assetsPath: assetsPath,
		categories: []CategoryInfo{},
	}

	if err := gen.loadAssets(); err != nil {
		return nil, fmt.Errorf("failed to load assets: %w", err)
	}

	return gen, nil
}

// loadAssets scans the assets directory for gopherize.me category folders
// Categories are numbered folders like "000-Body", "010-Eyes", "020-Mouth", etc.
func (g *Generator) loadAssets() error {
	// If assets directory doesn't exist, return error (user needs to download artwork)
	if _, err := os.Stat(g.assetsPath); os.IsNotExist(err) {
		return fmt.Errorf("assets directory does not exist: %s. Please download gopherize.me artwork", g.assetsPath)
	}

	// Pattern to match numbered category folders: "000-CategoryName" or "010-CategoryName"
	categoryPattern := regexp.MustCompile(`^(\d+)-(.+)$`)

	entries, err := os.ReadDir(g.assetsPath)
	if err != nil {
		return fmt.Errorf("failed to read assets directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if folder matches gopherize.me category pattern
		matches := categoryPattern.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue // Skip folders that don't match the pattern
		}

		order, err := strconv.Atoi(matches[1])
		if err != nil {
			continue // Skip if order number is invalid
		}

		categoryName := matches[2]
		categoryPath := filepath.Join(g.assetsPath, entry.Name())

		// Scan category folder for PNG files
		features, err := g.scanCategoryForFeatures(categoryPath)
		if err != nil {
			continue // Skip categories with errors
		}

		if len(features) > 0 {
			categoryType := g.detectCategoryType(categoryName)
			g.categories = append(g.categories, CategoryInfo{
				Order:    order,
				Name:     categoryName,
				Type:     categoryType,
				Features: features,
			})
		}
	}

	// Sort categories by order number
	sort.Slice(g.categories, func(i, j int) bool {
		return g.categories[i].Order < g.categories[j].Order
	})

	return nil
}

// scanCategoryForFeatures scans a category folder for PNG feature files
func (g *Generator) scanCategoryForFeatures(categoryPath string) ([]string, error) {
	features := []string{}

	entries, err := os.ReadDir(categoryPath)
	if err != nil {
		return features, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
			continue
		}

		fullPath := filepath.Join(categoryPath, entry.Name())
		features = append(features, fullPath)
	}

	return features, nil
}

// detectCategoryType determines the type of category based on its name
func (g *Generator) detectCategoryType(name string) CategoryType {
	lower := strings.ToLower(name)
	
	switch {
	case strings.Contains(lower, "body"):
		return CategoryTypeBody
	case strings.Contains(lower, "eye"):
		return CategoryTypeEyes
	case strings.Contains(lower, "shirt"):
		return CategoryTypeShirts
	case strings.Contains(lower, "hair") && !strings.Contains(lower, "facial"):
		return CategoryTypeHair
	case strings.Contains(lower, "facial") || strings.Contains(lower, "beard") || strings.Contains(lower, "mustache") || strings.Contains(lower, "stache"):
		return CategoryTypeFacialHair
	case strings.Contains(lower, "glass"):
		return CategoryTypeGlasses
	case strings.Contains(lower, "hat") || strings.Contains(lower, "accessor"):
		return CategoryTypeHats
	case strings.Contains(lower, "extra"):
		return CategoryTypeExtras
	default:
		return CategoryTypeUnknown
	}
}

// getCategoriesByType returns all categories of a specific type
func (g *Generator) getCategoriesByType(categoryType CategoryType) []CategoryInfo {
	var result []CategoryInfo
	for _, cat := range g.categories {
		if cat.Type == categoryType {
			result = append(result, cat)
		}
	}
	return result
}

// getCategoryOrderFromPath extracts the category order number from a file path
func (g *Generator) getCategoryOrderFromPath(path string) int {
	// Path format: "assets/artwork/010-Body/feature.png"
	// Extract the order number (010) from the folder name
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if strings.Contains(part, "-") {
			// Check if it matches the pattern "010-CategoryName"
			re := regexp.MustCompile(`^(\d+)-`)
			matches := re.FindStringSubmatch(part)
			if len(matches) >= 2 {
				if order, err := strconv.Atoi(matches[1]); err == nil {
					return order
				}
			}
		}
	}
	return 999 // Default to high number if not found (will be sorted last)
}

// isRareFeature determines if a feature is rare based on filename
func (g *Generator) isRareFeature(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.Contains(lower, "rare") ||
		strings.Contains(lower, "legendary") ||
		strings.Contains(lower, "epic") ||
		strings.Contains(lower, "gold") ||
		strings.Contains(lower, "diamond")
}

// GenerateOptions configures sprite generation
type GenerateOptions struct {
	Complexity    int
	TargetRarity  string
	Seed          int64
	PreserveLayers []string // Layer IDs to preserve (for evolution)
}

// GenerateResult contains the generated sprite and metadata
type GenerateResult struct {
	Image         image.Image
	Layers        []string // List of layer file paths used
	Complexity    int
	Rarity        string
}

// Generate creates a procedurally generated gopher sprite using gopherize.me categories
// Follows rarity-based rules:
// - ALL: Body + Eyes + 1 Extra
// - UNCOMMON: + Facial Hair OR Glasses
// - RARE: + Hair OR Facial Hair + Another Extra OR Glasses
// - LEGENDARY: All categories + 3 Extras
func (g *Generator) Generate(opts GenerateOptions) (*GenerateResult, error) {
	if len(g.categories) == 0 {
		return nil, fmt.Errorf("no categories loaded - please ensure artwork is downloaded")
	}

	rng := rand.New(rand.NewSource(opts.Seed))
	if opts.Seed == 0 {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	// Determine target rarity
	targetRarity := opts.TargetRarity
	if targetRarity == "" {
		// If complexity is specified, convert to rarity
		if opts.Complexity > 0 {
			targetRarity = g.complexityToRarity(opts.Complexity)
		} else {
			// Random rarity for wild encounters (weighted distribution)
			randFloat := rng.Float64()
			switch {
			case randFloat < 0.60:
				targetRarity = "COMMON"
			case randFloat < 0.85:
				targetRarity = "UNCOMMON"
			case randFloat < 0.95:
				targetRarity = "RARE"
			case randFloat < 0.99:
				targetRarity = "EPIC"
			default:
				targetRarity = "LEGENDARY"
			}
		}
	}

	// Separate lists for non-extra features and extra features
	// Extras will be composited last to appear on top
	nonExtraFeatures := []string{} // Features to composite first (in order)
	extraFeatures := []string{}    // Extras to composite last (on top)
	usedCategoryIndices := make(map[int]bool)

	// Get category indices by type for easy lookup
	bodyCats := g.getCategoriesByType(CategoryTypeBody)
	eyesCats := g.getCategoriesByType(CategoryTypeEyes)
	hairCats := g.getCategoriesByType(CategoryTypeHair)
	facialHairCats := g.getCategoriesByType(CategoryTypeFacialHair)
	glassesCats := g.getCategoriesByType(CategoryTypeGlasses)
	extrasCats := g.getCategoriesByType(CategoryTypeExtras)

	// Find category indices in main categories array
	findCategoryIndex := func(cat CategoryInfo) int {
		for i, c := range g.categories {
			if c.Order == cat.Order && c.Name == cat.Name {
				return i
			}
		}
		return -1
	}

	// Helper to add a feature from a category (returns feature path)
	addFeatureFromCategory := func(cat CategoryInfo) (string, error) {
		if len(cat.Features) == 0 {
			return "", fmt.Errorf("category has no features")
		}
		feature := cat.Features[rng.Intn(len(cat.Features))]
		idx := findCategoryIndex(cat)
		if idx >= 0 {
			usedCategoryIndices[idx] = true
		}
		return feature, nil
	}

	// Step 1: ALL gophers get Body + Eyes + 1 Extra
	if len(bodyCats) == 0 {
		return nil, fmt.Errorf("no body categories found")
	}
	bodyCat := bodyCats[rng.Intn(len(bodyCats))]
	bodyFeature, err := addFeatureFromCategory(bodyCat)
	if err != nil {
		return nil, fmt.Errorf("failed to get body feature: %w", err)
	}
	nonExtraFeatures = append(nonExtraFeatures, bodyFeature)

	// Add eyes
	if len(eyesCats) == 0 {
		return nil, fmt.Errorf("no eyes categories found")
	}
	eyesCat := eyesCats[rng.Intn(len(eyesCats))]
	eyesFeature, err := addFeatureFromCategory(eyesCat)
	if err != nil {
		return nil, fmt.Errorf("failed to get eyes feature: %w", err)
	}
	nonExtraFeatures = append(nonExtraFeatures, eyesFeature)

	// Add 1 Extra (required for all) - will be composited last
	if len(extrasCats) == 0 {
		return nil, fmt.Errorf("no extras categories found")
	}
	extraCat := extrasCats[rng.Intn(len(extrasCats))]
	extraFeature, err := addFeatureFromCategory(extraCat)
	if err != nil {
		return nil, fmt.Errorf("failed to get extra feature: %w", err)
	}
	extraFeatures = append(extraFeatures, extraFeature)

	// Step 2: Apply rarity-based rules (collect features, don't composite yet)
	switch targetRarity {
	case "UNCOMMON":
		// UNCOMMON: + Facial Hair OR Glasses
		if rng.Float32() < 0.5 && len(facialHairCats) > 0 {
			// Add facial hair
			facialHairCat := facialHairCats[rng.Intn(len(facialHairCats))]
			feature, err := addFeatureFromCategory(facialHairCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		} else if len(glassesCats) > 0 {
			// Add glasses
			glassesCat := glassesCats[rng.Intn(len(glassesCats))]
			feature, err := addFeatureFromCategory(glassesCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		}

	case "RARE":
		// RARE: + (Hair OR Facial Hair) AND (Another Extra OR Glasses)
		// Add hair OR facial hair (always add one)
		if rng.Float32() < 0.5 && len(hairCats) > 0 {
			hairCat := hairCats[rng.Intn(len(hairCats))]
			feature, err := addFeatureFromCategory(hairCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		} else if len(facialHairCats) > 0 {
			facialHairCat := facialHairCats[rng.Intn(len(facialHairCats))]
			feature, err := addFeatureFromCategory(facialHairCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		}

		// Add another extra OR glasses (always add one)
		if rng.Float32() < 0.5 && len(extrasCats) > 0 {
			// Find unused extra category
			availableExtras := []CategoryInfo{}
			for _, cat := range extrasCats {
				idx := findCategoryIndex(cat)
				if idx >= 0 && !usedCategoryIndices[idx] {
					availableExtras = append(availableExtras, cat)
				}
			}
			// If no unused extras, allow reusing
			if len(availableExtras) == 0 {
				availableExtras = extrasCats
			}
			if len(availableExtras) > 0 {
				extraCat := availableExtras[rng.Intn(len(availableExtras))]
				feature, err := addFeatureFromCategory(extraCat)
				if err == nil {
					extraFeatures = append(extraFeatures, feature)
				}
			}
		} else if len(glassesCats) > 0 {
			glassesCat := glassesCats[rng.Intn(len(glassesCats))]
			feature, err := addFeatureFromCategory(glassesCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		}

	case "EPIC":
		// EPIC: Similar to RARE but with more features
		// Add hair OR facial hair
		if rng.Float32() < 0.5 && len(hairCats) > 0 {
			hairCat := hairCats[rng.Intn(len(hairCats))]
			feature, err := addFeatureFromCategory(hairCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		} else if len(facialHairCats) > 0 {
			facialHairCat := facialHairCats[rng.Intn(len(facialHairCats))]
			feature, err := addFeatureFromCategory(facialHairCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		}

		// Add glasses
		if len(glassesCats) > 0 {
			glassesCat := glassesCats[rng.Intn(len(glassesCats))]
			feature, err := addFeatureFromCategory(glassesCat)
			if err == nil {
				nonExtraFeatures = append(nonExtraFeatures, feature)
			}
		}

		// Add another extra
		availableExtras := []CategoryInfo{}
		for _, cat := range extrasCats {
			idx := findCategoryIndex(cat)
			if idx >= 0 && !usedCategoryIndices[idx] {
				availableExtras = append(availableExtras, cat)
			}
		}
		if len(availableExtras) > 0 {
			extraCat := availableExtras[rng.Intn(len(availableExtras))]
			feature, err := addFeatureFromCategory(extraCat)
			if err == nil {
				extraFeatures = append(extraFeatures, feature)
			}
		}

	case "LEGENDARY":
		// LEGENDARY: All categories + 3 Extras
		// Add all available categories in order (categories are already sorted by order)
		for i, category := range g.categories {
			if usedCategoryIndices[i] {
				continue
			}
			if len(category.Features) == 0 {
				usedCategoryIndices[i] = true
				continue
			}

			// Skip extras for now (we'll add 3 at the end)
			if category.Type == CategoryTypeExtras {
				continue
			}

			feature := category.Features[rng.Intn(len(category.Features))]
			nonExtraFeatures = append(nonExtraFeatures, feature)
			usedCategoryIndices[i] = true
		}

		// Add 3 more extras (can reuse the same extra category if needed)
		// First try to find unused extra categories
		availableExtras := []CategoryInfo{}
		for _, cat := range extrasCats {
			idx := findCategoryIndex(cat)
			if idx >= 0 && !usedCategoryIndices[idx] {
				availableExtras = append(availableExtras, cat)
			}
		}
		// If we don't have enough unique extra categories, allow reusing
		if len(availableExtras) == 0 {
			availableExtras = extrasCats
		}
		
		for count := 0; count < 3; count++ {
			if len(availableExtras) > 0 {
				extraCat := availableExtras[rng.Intn(len(availableExtras))]
				feature := extraCat.Features[rng.Intn(len(extraCat.Features))]
				extraFeatures = append(extraFeatures, feature)
			}
		}
	}

	// Step 3: Composite all features in the correct order
	// Sort non-extra features by their category order to ensure proper layering
	sort.Slice(nonExtraFeatures, func(i, j int) bool {
		// Extract category order from file path (e.g., "assets/artwork/010-Body/feature.png")
		orderI := g.getCategoryOrderFromPath(nonExtraFeatures[i])
		orderJ := g.getCategoryOrderFromPath(nonExtraFeatures[j])
		return orderI < orderJ
	})

	// First, composite all non-extra features in order
	baseImg, err := g.loadImage(nonExtraFeatures[0])
	if err != nil {
		return nil, fmt.Errorf("failed to load base image: %w", err)
	}

	// Composite remaining non-extra features in order
	for i := 1; i < len(nonExtraFeatures); i++ {
		img, err := g.loadImage(nonExtraFeatures[i])
		if err == nil {
			baseImg = g.compositeLayer(baseImg, img)
		}
	}

	// Finally, composite all extras on top (extras appear on top of everything)
	for _, extraFeature := range extraFeatures {
		img, err := g.loadImage(extraFeature)
		if err == nil {
			baseImg = g.compositeLayer(baseImg, img)
		}
	}

	// Combine all layers for the result
	layers := append(nonExtraFeatures, extraFeatures...)

	// Calculate complexity based on number of layers
	complexity := len(layers)
	
	// Use target rarity since we built the gopher to match it
	rarity := targetRarity

	return &GenerateResult{
		Image:      baseImg,
		Layers:     layers,
		Complexity: complexity,
		Rarity:     rarity,
	}, nil
}

// compositeLayer overlays one image on top of another
func (g *Generator) compositeLayer(base, overlay image.Image) image.Image {
	bounds := base.Bounds()
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, base, image.Point{}, draw.Src)
	draw.Draw(result, bounds, overlay, image.Point{}, draw.Over)
	return result
}

// loadImage loads a PNG image from file
func (g *Generator) loadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// SaveImage saves an image to a file
func (g *Generator) SaveImage(img image.Image, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// EncodeImageToBase64 encodes an image to base64 string
func (g *Generator) EncodeImageToBase64(img image.Image) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DecodeImageFromBase64 decodes a base64 string to an image
func (g *Generator) DecodeImageFromBase64(base64Str string) (image.Image, error) {
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}
	return img, nil
}

// LoadImageFromPath loads an image from a file path (for backward compatibility)
func (g *Generator) LoadImageFromPath(path string) (image.Image, error) {
	return g.loadImage(path)
}

// InvertColors inverts the colors of an image (for shiny gophers)
func (g *Generator) InvertColors(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := img.At(x, y)
			r, g, b, a := originalColor.RGBA()
			
			// Invert RGB values (keep alpha channel)
			// RGBA returns values in 0-65535 range, so we need to convert
			invertedR := uint8(255 - (r >> 8))
			invertedG := uint8(255 - (g >> 8))
			invertedB := uint8(255 - (b >> 8))
			invertedA := uint8(a >> 8)
			
			result.Set(x, y, color.RGBA{
				R: invertedR,
				G: invertedG,
				B: invertedB,
				A: invertedA,
			})
		}
	}
	
	return result
}

// AddShinyGlow adds a glowing effect behind a shiny gopher (optimized for performance)
func (g *Generator) AddShinyGlow(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	// Create a larger canvas to accommodate the glow
	glowSize := 25 // Extra pixels for glow on each side
	canvasWidth := width + glowSize*2
	canvasHeight := height + glowSize*2
	
	// Create the final result with transparent background
	result := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	
	// Fast glow: create a simple expanded outline
	// Find edges of the gopher (where alpha > 0)
	edgePixels := make([]struct{ x, y int }, 0)
	
	// Sample every 3rd pixel for speed
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 3 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 3 {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				edgePixels = append(edgePixels, struct{ x, y int }{x, y})
			}
		}
	}
	
	// Create glow by drawing expanded circles around edge pixels
	glowColor := color.RGBA{R: 255, G: 255, B: 200, A: 150} // Golden glow
	glowRadius := 8
	
	for _, pixel := range edgePixels {
		// Draw glow circles around each edge pixel
		for dy := -glowRadius; dy <= glowRadius; dy++ {
			for dx := -glowRadius; dx <= glowRadius; dx++ {
				distSq := dx*dx + dy*dy
				if distSq <= glowRadius*glowRadius {
					// Calculate opacity based on distance (fade out)
					opacity := uint8(float64(glowColor.A) * (1.0 - float64(distSq)/float64(glowRadius*glowRadius)))
					
					glowX := glowSize + pixel.x + dx
					glowY := glowSize + pixel.y + dy
					
					if glowX >= 0 && glowX < canvasWidth && glowY >= 0 && glowY < canvasHeight {
						// Blend with existing glow - use max alpha
						existing := result.At(glowX, glowY)
						_, _, _, ea := existing.RGBA()
						
						// Use max alpha for blending
						existingAlpha := uint8(ea >> 8)
						newA := opacity
						if existingAlpha > newA {
							newA = existingAlpha
						}
						result.Set(glowX, glowY, color.RGBA{
							R: glowColor.R,
							G: glowColor.G,
							B: glowColor.B,
							A: newA,
						})
					}
				}
			}
		}
	}
	
	// Draw the original inverted image on top
	originalX := glowSize
	originalY := glowSize
	draw.Draw(result, image.Rect(originalX, originalY, originalX+width, originalY+height),
		img, bounds.Min, draw.Over)
	
	return result
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// complexityToRarity converts complexity score to rarity string
func (g *Generator) complexityToRarity(complexity int) string {
	switch {
	case complexity <= 2:
		return "COMMON"
	case complexity <= 4:
		return "UNCOMMON"
	case complexity <= 6:
		return "RARE"
	case complexity <= 8:
		return "EPIC"
	default:
		return "LEGENDARY"
	}
}

