package game

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"gophermon-bot/internal/gopherkon"
)

// EvolutionService handles gopher evolution
type EvolutionService struct {
	generator *gopherkon.Generator
	assetsPath string
}

func NewEvolutionService(generator *gopherkon.Generator, assetsPath string) *EvolutionService {
	return &EvolutionService{
		generator:  generator,
		assetsPath: assetsPath,
	}
}

// CheckEvolution checks if a gopher should evolve and performs evolution
func (es *EvolutionService) EvolveGopher(gopher *Gopher) (*Gopher, error) {
	// Evolution thresholds: level 16 and 32
	shouldEvolve := false
	newStage := gopher.EvolutionStage

	if gopher.Level >= 16 && gopher.EvolutionStage == 0 {
		shouldEvolve = true
		newStage = 1
	} else if gopher.Level >= 32 && gopher.EvolutionStage == 1 {
		shouldEvolve = true
		newStage = 2
	}

	if !shouldEvolve {
		return nil, nil
	}

	// Preserve some layers from base gopher
	preserveCount := 1 + rand.Intn(2) // Preserve 1-2 layers
	if preserveCount > len(gopher.GopherkonLayers) {
		preserveCount = len(gopher.GopherkonLayers)
	}

	preservedLayers := []string{}
	if preserveCount > 0 {
		preservedLayers = gopher.GopherkonLayers[:preserveCount]
	}

	// Generate new sprite with increased complexity
	currentComplexity := gopher.ComplexityScore
	newComplexity := currentComplexity + 2 + rand.Intn(3) // Add 2-4 complexity

	// Ensure rarity upgrade
	currentRarity := ComplexityToRarity(currentComplexity)
	targetRarity := ComplexityToRarity(newComplexity)
	
	// Force at least one tier upgrade if not already high
	if targetRarity == currentRarity && currentRarity != RarityLegendary {
		newComplexity = currentComplexity + 3
		targetRarity = ComplexityToRarity(newComplexity)
	}

	// Generate evolution sprite
	result, err := es.generator.Generate(gopherkon.GenerateOptions{
		Complexity:    newComplexity,
		TargetRarity:  targetRarity.String(),
		Seed:         time.Now().UnixNano(),
		PreserveLayers: preservedLayers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate evolution sprite: %w", err)
	}

	// Save new sprite
	spritePath := filepath.Join("assets", "generated", fmt.Sprintf("evolved_%s_%d.png", gopher.ID, time.Now().Unix()))
	if err := es.generator.SaveImage(result.Image, spritePath); err != nil {
		return nil, fmt.Errorf("failed to save evolution sprite: %w", err)
	}

	// Update gopher
	gopher.EvolutionStage = newStage
	gopher.SpritePath = spritePath
	gopher.ComplexityScore = result.Complexity
	gopher.Rarity = result.Rarity
	gopher.GopherkonLayers = result.Layers

	// Stat boost on evolution
	hpBoost := 20 + (newStage * 10)
	gopher.MaxHP += hpBoost
	gopher.CurrentHP += hpBoost // Heal on evolution
	gopher.Attack += 5 + (newStage * 3)
	gopher.Defense += 5 + (newStage * 3)
	gopher.Speed += 3 + (newStage * 2)

	// Unlock new abilities if level is high enough
	if gopher.Level >= 20 && len(gopher.Abilities) < 4 {
		abilityTemplates := GetAbilitiesForArchetype(Archetype(gopher.SpeciesArchetype))
		if len(abilityTemplates) > len(gopher.Abilities) {
			// Add a new ability
			newAbilityTemplate := abilityTemplates[len(gopher.Abilities)]
			newAbility, err := CreateAbilityFromTemplate(newAbilityTemplate, fmt.Sprintf("%s_ability_%d", gopher.ID, len(gopher.Abilities)))
			if err == nil {
				gopher.Abilities = append(gopher.Abilities, newAbility)
			}
		}
	}

	return gopher, nil
}

// CheckAndEvolve checks if evolution should occur after level up
func (es *EvolutionService) CheckAndEvolve(gopher *Gopher) (evolved bool, evolutionMessage string) {
	oldStage := gopher.EvolutionStage
	evolvedGopher, err := es.EvolveGopher(gopher)
	if err != nil || evolvedGopher == nil {
		return false, ""
	}

	if evolvedGopher.EvolutionStage > oldStage {
		evolutionMessage = fmt.Sprintf("🎉 **%s is evolving!** 🎉\n", gopher.Name)
		evolutionMessage += fmt.Sprintf("Evolution stage: %d → %d\n", oldStage, evolvedGopher.EvolutionStage)
		evolutionMessage += fmt.Sprintf("Rarity: %s → %s\n", gopher.Rarity, evolvedGopher.Rarity)
		evolutionMessage += fmt.Sprintf("Stats increased significantly!")
		return true, evolutionMessage
	}

	return false, ""
}

