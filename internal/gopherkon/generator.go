package gopherkon

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// LayerType represents different types of sprite layers
type LayerType string

const (
	LayerBody       LayerType = "body"
	LayerEars       LayerType = "ears"
	LayerEyes       LayerType = "eyes"
	LayerMouth      LayerType = "mouth"
	LayerAccessory  LayerType = "accessory"
	LayerHat        LayerType = "hat"
	LayerClothing   LayerType = "clothing"
	LayerTool       LayerType = "tool"
)

// AssetInfo represents information about a gopherkon asset
type AssetInfo struct {
	Path      string
	LayerType LayerType
	IsRare    bool
}

// Generator handles sprite generation from gopherkon assets
type Generator struct {
	assetsPath string
	assets     map[LayerType][]AssetInfo
}

// NewGenerator creates a new sprite generator
func NewGenerator(assetsPath string) (*Generator, error) {
	gen := &Generator{
		assetsPath: assetsPath,
		assets:     make(map[LayerType][]AssetInfo),
	}

	if err := gen.loadAssets(); err != nil {
		return nil, fmt.Errorf("failed to load assets: %w", err)
	}

	return gen, nil
}

// loadAssets scans the assets directory and categorizes sprites
func (g *Generator) loadAssets() error {
	// If assets directory doesn't exist, create a placeholder structure
	if _, err := os.Stat(g.assetsPath); os.IsNotExist(err) {
		// Create directory structure
		os.MkdirAll(filepath.Join(g.assetsPath, "body"), 0755)
		os.MkdirAll(filepath.Join(g.assetsPath, "eyes"), 0755)
		os.MkdirAll(filepath.Join(g.assetsPath, "mouth"), 0755)
		os.MkdirAll(filepath.Join(g.assetsPath, "hats"), 0755)
		os.MkdirAll(filepath.Join(g.assetsPath, "accessories"), 0755)
		return nil
	}

	// Scan for asset files - map to actual gopherkon directory structure
	layerDirs := map[LayerType]string{
		LayerBody:      "torso",      // gopherkon uses "torso" for body
		LayerEars:      "ears",       // ears directory
		LayerEyes:      "eyes",
		LayerMouth:     "mouth",
		LayerHat:       "hat",
		LayerAccessory: "accessory",
		LayerClothing:  "extras",     // Use extras for additional accessories
		LayerTool:      "pose",      // Use pose for tools/items
	}

	for layerType, dirName := range layerDirs {
		dirPath := filepath.Join(g.assetsPath, dirName)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}

		// Recursively scan directory for PNG files (handles subdirectories like torso/normal/)
		g.scanDirectoryForAssets(dirPath, layerType)
	}

	return nil
}

// scanDirectoryForAssets recursively scans a directory for PNG assets
func (g *Generator) scanDirectoryForAssets(dirPath string, layerType LayerType) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		
		if entry.IsDir() {
			// Recursively scan subdirectories
			g.scanDirectoryForAssets(fullPath, layerType)
			continue
		}

		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".png") {
			continue
		}

		// Skip "none.png" files
		if strings.Contains(strings.ToLower(entry.Name()), "none.png") {
			continue
		}

		isRare := strings.Contains(strings.ToLower(entry.Name()), "rare") ||
			strings.Contains(strings.ToLower(entry.Name()), "legendary") ||
			strings.Contains(strings.ToLower(entry.Name()), "epic") ||
			strings.Contains(strings.ToLower(entry.Name()), "santa") ||
			strings.Contains(strings.ToLower(entry.Name()), "tophat") ||
			strings.Contains(strings.ToLower(entry.Name()), "sherlock")

		g.assets[layerType] = append(g.assets[layerType], AssetInfo{
			Path:      fullPath,
			LayerType: layerType,
			IsRare:    isRare,
		})
	}
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

// Generate creates a procedurally generated gopher sprite
func (g *Generator) Generate(opts GenerateOptions) (*GenerateResult, error) {
	rng := rand.New(rand.NewSource(opts.Seed))
	if opts.Seed == 0 {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}

	layers := []string{}
	complexity := 0

	// Always start with a body (base layer) - fully randomized
	bodyAssets := g.assets[LayerBody]
	if len(bodyAssets) == 0 {
		return nil, fmt.Errorf("no body assets available")
	}
	
	// Fully random body selection
	bodyAsset := bodyAssets[rng.Intn(len(bodyAssets))]
	layers = append(layers, bodyAsset.Path)
	complexity++

	// Load base image
	baseImg, err := g.loadImage(bodyAsset.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load body: %w", err)
	}

	// Determine target complexity
	targetComplexity := opts.Complexity
	if targetComplexity == 0 {
		// Random complexity for wild encounters
		targetComplexity = 2 + rng.Intn(8) // 2-9
	}

	// ALWAYS add ears (required)
	earsAssets := g.assets[LayerEars]
	if len(earsAssets) > 0 {
		// Random ear style (fancy, fluffy, foxy, normal, playful, pointy, tiny, wide, wolf)
		earAsset := earsAssets[rng.Intn(len(earsAssets))]
		layers = append(layers, earAsset.Path)
		complexity++
		if earAsset.IsRare {
			complexity++ // Rare assets add extra complexity
		}

		earImg, err := g.loadImage(earAsset.Path)
		if err == nil {
			baseImg = g.compositeLayer(baseImg, earImg)
		}
	}

	// Add eyes (randomized)
	eyesAssets := g.assets[LayerEyes]
	if len(eyesAssets) > 0 {
		eyeAsset := eyesAssets[rng.Intn(len(eyesAssets))]
		layers = append(layers, eyeAsset.Path)
		complexity++
		if eyeAsset.IsRare {
			complexity++ // Rare assets add extra complexity
		}

		eyeImg, err := g.loadImage(eyeAsset.Path)
		if err == nil {
			baseImg = g.compositeLayer(baseImg, eyeImg)
		}
	}

	// Add mouth (randomized)
	mouthAssets := g.assets[LayerMouth]
	if len(mouthAssets) > 0 {
		mouthAsset := mouthAssets[rng.Intn(len(mouthAssets))]
		layers = append(layers, mouthAsset.Path)
		complexity++
		if mouthAsset.IsRare {
			complexity++
		}

		mouthImg, err := g.loadImage(mouthAsset.Path)
		if err == nil {
			baseImg = g.compositeLayer(baseImg, mouthImg)
		}
	}

	// Add accessories up to target complexity (fully randomized order)
	accessoryTypes := []LayerType{LayerHat, LayerAccessory, LayerClothing, LayerTool}
	// Shuffle accessory types for randomization
	for i := len(accessoryTypes) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		accessoryTypes[i], accessoryTypes[j] = accessoryTypes[j], accessoryTypes[i]
	}
	
	for complexity < targetComplexity && len(accessoryTypes) > 0 {
		// Try each accessory type in random order
		added := false
		for _, accType := range accessoryTypes {
			accAssets := g.assets[accType]
			if len(accAssets) == 0 {
				continue
			}

			// Random asset from this type
			accAsset := accAssets[rng.Intn(len(accAssets))]
			layers = append(layers, accAsset.Path)
			complexity++
			if accAsset.IsRare {
				complexity++
			}

			accImg, err := g.loadImage(accAsset.Path)
			if err == nil {
				baseImg = g.compositeLayer(baseImg, accImg)
				added = true
				break
			}
		}
		
		// If we couldn't add any accessory, break to avoid infinite loop
		if !added {
			break
		}
	}

	rarity := g.complexityToRarity(complexity)

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

