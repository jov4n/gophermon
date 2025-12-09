package game

import (
	"fmt"
	"math/rand"
	"time"

	"gophermon-bot/internal/gopherkon"
)

// EvolutionService handles gopher evolution
type EvolutionService struct {
	generator *gopherkon.Generator
	assetsPath string
	eventManager *EventManager
}

func NewEvolutionService(generator *gopherkon.Generator, assetsPath string) *EvolutionService {
	return &EvolutionService{
		generator:  generator,
		assetsPath: assetsPath,
		eventManager: nil, // Will be set by service
	}
}

// SetEventManager sets the event manager for evolution level reduction
func (es *EvolutionService) SetEventManager(em *EventManager) {
	es.eventManager = em
}

// CheckEvolution checks if a gopher should evolve and performs evolution
func (es *EvolutionService) EvolveGopher(gopher *Gopher) (*Gopher, error) {
	// Evolution thresholds: level 16 and 32 (reduced by events)
	levelReduction := 0
	if es.eventManager != nil {
		levelReduction = es.eventManager.GetEvolutionLevelReduction()
	}
	
	threshold1 := 16 - levelReduction
	threshold2 := 32 - levelReduction
	
	if threshold1 < 1 {
		threshold1 = 1
	}
	if threshold2 < threshold1+1 {
		threshold2 = threshold1 + 1
	}
	
	shouldEvolve := false
	newStage := gopher.EvolutionStage

	if gopher.Level >= threshold1 && gopher.EvolutionStage == 0 {
		shouldEvolve = true
		newStage = 1
	} else if gopher.Level >= threshold2 && gopher.EvolutionStage == 1 {
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

	// Encode new sprite to base64
	spriteData, err := es.generator.EncodeImageToBase64(result.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to encode evolution sprite: %w", err)
	}

	// Update gopher
	gopher.EvolutionStage = newStage
	gopher.SpritePath = "" // No longer using file paths
	gopher.SpriteData = spriteData
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

	// Unlock new abilities based on evolution stage
	abilityTemplates := GetAbilitiesForArchetype(Archetype(gopher.SpeciesArchetype), gopher.EvolutionStage, gopher.Rarity)
	
	// Determine how many abilities gopher should have
	numAbilities := 2
	if gopher.Level >= 10 {
		numAbilities = 3
	}
	if gopher.Level >= 20 {
		numAbilities = 4
	}
	if gopher.EvolutionStage >= 1 {
		numAbilities = 5
	}
	if gopher.EvolutionStage >= 2 {
		numAbilities = 6
	}
	if gopher.Rarity == "LEGENDARY" {
		numAbilities = 7 // Legendaries get more abilities
	}
	
	// Add new abilities if needed
	if len(gopher.Abilities) < numAbilities && len(abilityTemplates) > len(gopher.Abilities) {
		// Add abilities starting from where we left off
		for i := len(gopher.Abilities); i < numAbilities && i < len(abilityTemplates); i++ {
			newAbility, err := CreateAbilityFromTemplate(abilityTemplates[i], fmt.Sprintf("%s_ability_%d", gopher.ID, i))
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

