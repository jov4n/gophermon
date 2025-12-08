package main

import (
	"flag"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"strings"
	"time"

	"gophermon-bot/internal/gopherkon"
)

func main() {
	// Command-line flags
	var (
		rarity     = flag.String("rarity", "", "Rarity: COMMON, UNCOMMON, RARE, EPIC, LEGENDARY (empty = random)")
		complexity = flag.Int("complexity", -1, "Complexity: 1-15 (-1 = based on rarity or random)")
		seed       = flag.Int64("seed", -1, "Random seed (-1 = use current time)")
		output     = flag.String("output", "test_artwork.png", "Output file for sprite")
		count      = flag.Int("count", 1, "Number of sprites to generate")
		showInfo   = flag.Bool("info", true, "Show generation info")
	)
	flag.Parse()

	// Set random seed
	if *seed == -1 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)

	if *showInfo {
		fmt.Println("=== Artwork Generation Test ===")
		fmt.Printf("Seed: %d\n", *seed)
		fmt.Printf("Generating %d sprite(s)...\n\n", *count)
	}

	// Initialize generator
	assetsPath := "assets/artwork"
	generator, err := gopherkon.NewGenerator(assetsPath)
	if err != nil {
		fmt.Printf("Error loading generator: %v\n", err)
		return
	}

	// Generate sprites
	for i := 0; i < *count; i++ {
		// Determine rarity
		var selectedRarity string
		if *rarity != "" {
			selectedRarity = strings.ToUpper(*rarity)
			valid := false
			for _, r := range []string{"COMMON", "UNCOMMON", "RARE", "EPIC", "LEGENDARY"} {
				if selectedRarity == r {
					valid = true
					break
				}
			}
			if !valid {
				fmt.Printf("Invalid rarity: %s. Using random.\n", *rarity)
				selectedRarity = getRandomRarity()
			}
		} else {
			selectedRarity = getRandomRarity()
		}

		// Determine complexity
		var selectedComplexity int
		if *complexity == -1 {
			// Random complexity based on rarity
			selectedComplexity = getComplexityForRarity(selectedRarity)
		} else {
			selectedComplexity = *complexity
			if selectedComplexity < 1 {
				selectedComplexity = 1
			}
			if selectedComplexity > 15 {
				selectedComplexity = 15
			}
		}

		// Use different seed for each sprite if generating multiple
		currentSeed := *seed
		if *count > 1 {
			currentSeed = *seed + int64(i)
		}

		// Generate sprite
		result, err := generator.Generate(gopherkon.GenerateOptions{
			Complexity:   selectedComplexity,
			TargetRarity: selectedRarity,
			Seed:        currentSeed,
		})
		if err != nil {
			fmt.Printf("Error generating sprite %d: %v\n", i+1, err)
			continue
		}

		// Determine output filename
		outputFile := *output
		if *count > 1 {
			// Add index to filename
			if strings.HasSuffix(outputFile, ".png") {
				outputFile = strings.TrimSuffix(outputFile, ".png")
			}
			outputFile = fmt.Sprintf("%s_%d.png", outputFile, i+1)
		} else if !strings.HasSuffix(outputFile, ".png") {
			// Ensure .png extension for single file
			outputFile = outputFile + ".png"
		}

		// Save sprite
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", outputFile, err)
			continue
		}

		err = png.Encode(file, result.Image)
		file.Close()
		if err != nil {
			fmt.Printf("Error encoding PNG %s: %v\n", outputFile, err)
			continue
		}

		if *showInfo {
			fmt.Printf("✓ Generated: %s\n", outputFile)
			fmt.Printf("  Rarity: %s | Complexity: %d | Seed: %d\n", selectedRarity, selectedComplexity, currentSeed)
			fmt.Printf("  Dimensions: %dx%d\n", result.Image.Bounds().Dx(), result.Image.Bounds().Dy())
			fmt.Printf("  Layers: %d\n", len(result.Layers))
			if len(result.Layers) > 0 {
				fmt.Printf("  Layer list: %s\n", strings.Join(result.Layers, ", "))
			}
			fmt.Println()
		} else {
			fmt.Printf("Generated: %s\n", outputFile)
		}
	}

	if *showInfo {
		fmt.Println("=== Generation Complete ===")
		fmt.Println("\nUsage examples:")
		fmt.Println("  # Random artwork")
		fmt.Println("  go run scripts/test_artwork_generation.go")
		fmt.Println()
		fmt.Println("  # Specific rarity")
		fmt.Println("  go run scripts/test_artwork_generation.go -rarity=LEGENDARY")
		fmt.Println()
		fmt.Println("  # Specific complexity")
		fmt.Println("  go run scripts/test_artwork_generation.go -complexity=8")
		fmt.Println()
		fmt.Println("  # Custom output file")
		fmt.Println("  go run scripts/test_artwork_generation.go -output=my_sprite.png")
		fmt.Println()
		fmt.Println("  # Generate multiple sprites")
		fmt.Println("  go run scripts/test_artwork_generation.go -count=5 -rarity=RARE")
		fmt.Println()
		fmt.Println("  # Reproducible generation")
		fmt.Println("  go run scripts/test_artwork_generation.go -seed=12345")
	}
}

func getRandomRarity() string {
	rarities := []string{"COMMON", "UNCOMMON", "RARE", "EPIC", "LEGENDARY"}
	weights := []float64{0.60, 0.25, 0.10, 0.04, 0.01}
	
	r := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if r < cumulative {
			return rarities[i]
		}
	}
	return rarities[len(rarities)-1]
}

func getComplexityForRarity(rarity string) int {
	switch rarity {
	case "COMMON":
		return 1 + rand.Intn(2) // 1-2
	case "UNCOMMON":
		return 3 + rand.Intn(2) // 3-4
	case "RARE":
		return 5 + rand.Intn(2) // 5-6
	case "EPIC":
		return 7 + rand.Intn(2) // 7-8
	case "LEGENDARY":
		return 9 + rand.Intn(7) // 9-15
	default:
		return 1 + rand.Intn(15) // 1-15
	}
}

