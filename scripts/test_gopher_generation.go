package main

import (
	"flag"
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"strings"
	"time"

	"gophermon-bot/internal/game"
	"gophermon-bot/internal/gopherkon"
)

func main() {
	// Command-line flags
	var (
		archetype      = flag.String("archetype", "", "Archetype: Hacker, Tank, Speedy, Support, Mage (empty = random)")
		rarity         = flag.String("rarity", "", "Rarity: COMMON, UNCOMMON, RARE, EPIC, LEGENDARY (empty = random)")
		evolutionStage = flag.Int("evolution", -1, "Evolution stage: 0, 1, or 2 (-1 = random)")
		level          = flag.Int("level", -1, "Level: 1-100 (-1 = random 1-50)")
		complexity     = flag.Int("complexity", -1, "Complexity: 1-15 (-1 = based on rarity)")
		seed           = flag.Int64("seed", -1, "Random seed (-1 = use current time)")
		output         = flag.String("output", "", "Output file for sprite (empty = don't save)")
		showAbilities  = flag.Bool("abilities", true, "Show abilities list")
		showStats      = flag.Bool("stats", true, "Show stats")
		showStatus     = flag.Bool("status", true, "Show status effects info")
	)
	flag.Parse()

	// Set random seed
	if *seed == -1 {
		*seed = time.Now().UnixNano()
	}
	rand.Seed(*seed)

	fmt.Println("=== Gopher Generation Test ===")
	fmt.Printf("Seed: %d\n\n", *seed)

	// Determine archetype
	var selectedArchetype game.Archetype
	if *archetype != "" {
		selectedArchetype = game.Archetype(strings.Title(strings.ToLower(*archetype)))
		valid := false
		for _, a := range []game.Archetype{
			game.ArchetypeHacker,
			game.ArchetypeTank,
			game.ArchetypeSpeedy,
			game.ArchetypeSupport,
			game.ArchetypeMage,
		} {
			if selectedArchetype == a {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("Invalid archetype: %s. Using random.\n", *archetype)
			archetypes := []game.Archetype{
				game.ArchetypeHacker,
				game.ArchetypeTank,
				game.ArchetypeSpeedy,
				game.ArchetypeSupport,
				game.ArchetypeMage,
			}
			selectedArchetype = archetypes[rand.Intn(len(archetypes))]
		}
	} else {
		archetypes := []game.Archetype{
			game.ArchetypeHacker,
			game.ArchetypeTank,
			game.ArchetypeSpeedy,
			game.ArchetypeSupport,
			game.ArchetypeMage,
		}
		selectedArchetype = archetypes[rand.Intn(len(archetypes))]
	}

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
			selectedRarity = game.GetWildRarityDistribution(rand.Float64()).String()
		}
	} else {
		selectedRarity = game.GetWildRarityDistribution(rand.Float64()).String()
	}

	// Determine evolution stage
	var selectedEvolutionStage int
	if *evolutionStage == -1 {
		// Random evolution stage based on level
		if *level > 0 {
			if *level >= 32 {
				selectedEvolutionStage = rand.Intn(3) // 0, 1, or 2
			} else if *level >= 16 {
				selectedEvolutionStage = rand.Intn(2) // 0 or 1
			} else {
				selectedEvolutionStage = 0
			}
		} else {
			selectedEvolutionStage = rand.Intn(3) // 0, 1, or 2
		}
	} else {
		selectedEvolutionStage = *evolutionStage
		if selectedEvolutionStage < 0 || selectedEvolutionStage > 2 {
			fmt.Printf("Invalid evolution stage: %d. Using 0.\n", *evolutionStage)
			selectedEvolutionStage = 0
		}
	}

	// Determine level
	var selectedLevel int
	if *level == -1 {
		selectedLevel = 1 + rand.Intn(50)
	} else {
		selectedLevel = *level
		if selectedLevel < 1 {
			selectedLevel = 1
		}
		if selectedLevel > 100 {
			selectedLevel = 100
		}
	}

	// Determine complexity
	var selectedComplexity int
	if *complexity == -1 {
		// Determine complexity based on rarity
		min, max := game.RarityToComplexityRange(game.Rarity(selectedRarity))
		selectedComplexity = min + rand.Intn(max-min+1)
	} else {
		selectedComplexity = *complexity
		if selectedComplexity < 1 {
			selectedComplexity = 1
		}
		if selectedComplexity > 15 {
			selectedComplexity = 15
		}
	}

	// Generate sprite
	fmt.Println("Generating gopher sprite...")
	assetsPath := "assets/artwork"
	generator, err := gopherkon.NewGenerator(assetsPath)
	if err != nil {
		fmt.Printf("Error loading generator: %v\n", err)
		return
	}

	result, err := generator.Generate(gopherkon.GenerateOptions{
		Complexity:   selectedComplexity,
		TargetRarity: selectedRarity,
		Seed:        *seed,
	})
	if err != nil {
		fmt.Printf("Error generating sprite: %v\n", err)
		return
	}

	// Generate stats
	hp, attack, defense, speed := game.GenerateBaseStats(selectedArchetype, selectedRarity, selectedLevel)

	// Assign types
	primaryType := game.GetTypeFromArchetype(selectedArchetype)
	secondaryType := ""
	if rand.Float64() < 0.3 || selectedRarity == "RARE" || selectedRarity == "EPIC" || selectedRarity == "LEGENDARY" {
		secondaryType = string(game.GetRandomSecondaryType(primaryType))
	}

	// Create gopher name
	gopherName := game.GenerateGopherName(selectedArchetype)

	// Display information
	fmt.Println("\n=== Generated Gopher ===")
	fmt.Printf("Name: %s\n", gopherName)
	fmt.Printf("Archetype: %s\n", selectedArchetype)
	fmt.Printf("Rarity: %s\n", selectedRarity)
	fmt.Printf("Evolution Stage: %d\n", selectedEvolutionStage)
	fmt.Printf("Level: %d\n", selectedLevel)
	fmt.Printf("Complexity: %d\n", selectedComplexity)
	fmt.Printf("Primary Type: %s\n", primaryType)
	if secondaryType != "" {
		fmt.Printf("Secondary Type: %s\n", secondaryType)
	} else {
		fmt.Printf("Secondary Type: None\n")
	}

	if *showStats {
		fmt.Println("\n=== Stats ===")
		fmt.Printf("HP: %d / %d\n", hp, hp)
		fmt.Printf("Attack: %d\n", attack)
		fmt.Printf("Defense: %d\n", defense)
		fmt.Printf("Speed: %d\n", speed)
	}

	// Get abilities
	abilityTemplates := game.GetAbilitiesForArchetype(selectedArchetype, selectedEvolutionStage, selectedRarity)
	
	// Determine number of abilities based on level and evolution
	numAbilities := 2
	if selectedLevel >= 10 {
		numAbilities = 3
	}
	if selectedLevel >= 20 {
		numAbilities = 4
	}
	if selectedEvolutionStage >= 1 {
		numAbilities = 5
	}
	if selectedEvolutionStage >= 2 {
		numAbilities = 6
	}
	if selectedRarity == "LEGENDARY" {
		numAbilities = 7
	}

	if numAbilities > len(abilityTemplates) {
		numAbilities = len(abilityTemplates)
	}

	if *showAbilities {
		fmt.Println("\n=== Abilities ===")
		fmt.Printf("Total available: %d (showing first %d)\n\n", len(abilityTemplates), numAbilities)
		
		for i, templateID := range abilityTemplates[:numAbilities] {
			template, ok := game.AbilityTemplates[templateID]
			if !ok {
				fmt.Printf("%d. %s (template not found)\n", i+1, templateID)
				continue
			}
			
			// Determine if it's evolution-specific or legendary
			note := ""
			if strings.Contains(templateID, "concurrent") || strings.Contains(templateID, "mutex") || 
			   strings.Contains(templateID, "context") || strings.Contains(templateID, "reflect") || 
			   strings.Contains(templateID, "select") {
				note = " [Evolution Stage 1+]"
			} else if strings.Contains(templateID, "deadlock") || strings.Contains(templateID, "swarm") || 
			          strings.Contains(templateID, "overload") || strings.Contains(templateID, "recovery") || 
			          strings.Contains(templateID, "ultimate") {
				note = " [Evolution Stage 2+]"
			} else if strings.Contains(templateID, "legendary") || strings.Contains(templateID, "divine") || 
			          strings.Contains(templateID, "god") || strings.Contains(templateID, "apocalypse") || 
			          strings.Contains(templateID, "rewind") {
				note = " [LEGENDARY]"
			}
			
			fmt.Printf("%d. %s%s\n", i+1, template.Name, note)
			fmt.Printf("   Description: %s\n", template.Description)
			fmt.Printf("   Power: %d | Cost: %d | Target: %s\n", template.Power, template.Cost, template.Targeting)
			fmt.Println()
		}
		
		if len(abilityTemplates) > numAbilities {
			fmt.Printf("... and %d more abilities (unlock at higher levels/evolution)\n", len(abilityTemplates)-numAbilities)
		}
	}

	if *showStatus {
		fmt.Println("\n=== Status Effects Info ===")
		fmt.Println("Available status effects:")
		fmt.Println("  - Burn: Deals damage over time (12.5% max HP per turn)")
		fmt.Println("  - Poison: Deals damage over time (6.25% max HP + intensity per turn)")
		fmt.Println("  - Confusion: 33% chance to hurt self instead of attacking")
		fmt.Println("  - Paralysis: 25% chance to skip turn")
		fmt.Println("  - Sleep: Skip turn (30% chance to wake up each turn)")
		fmt.Println("  - Attack Up/Down: Modifies attack by 50%")
		fmt.Println("  - Defense Up/Down: Modifies defense by 50%")
		fmt.Println("  - Speed Up/Down: Modifies speed by 50%")
		fmt.Println("  - Protect: Blocks next attack")
	}

	// Save sprite if requested
	if *output != "" {
		fmt.Printf("\nSaving sprite to: %s\n", *output)
		file, err := os.Create(*output)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
		} else {
			defer file.Close()
			
			err = png.Encode(file, result.Image)
			if err != nil {
				fmt.Printf("Error encoding PNG: %v\n", err)
			} else {
				fmt.Printf("Sprite saved successfully!\n")
				fmt.Printf("Sprite dimensions: %dx%d\n", result.Image.Bounds().Dx(), result.Image.Bounds().Dy())
			}
		}
	}

	fmt.Println("\n=== Generation Complete ===")
	fmt.Println("\nUsage examples:")
	fmt.Println("  # Random gopher")
	fmt.Println("  go run scripts/test_gopher_generation.go")
	fmt.Println()
	fmt.Println("  # Specific archetype and rarity")
	fmt.Println("  go run scripts/test_gopher_generation.go -archetype=Hacker -rarity=LEGENDARY")
	fmt.Println()
	fmt.Println("  # Evolution stage 2 gopher")
	fmt.Println("  go run scripts/test_gopher_generation.go -evolution=2 -level=40")
	fmt.Println()
	fmt.Println("  # Save sprite to file")
	fmt.Println("  go run scripts/test_gopher_generation.go -output=test_gopher.png")
	fmt.Println()
	fmt.Println("  # Full custom gopher")
	fmt.Println("  go run scripts/test_gopher_generation.go -archetype=Mage -rarity=EPIC -evolution=1 -level=25 -complexity=7 -seed=12345")
}

