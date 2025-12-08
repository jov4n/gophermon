package game

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	"gophermon-bot/internal/gopherkon"
	"gophermon-bot/internal/storage"
)

type Service struct {
	trainerRepo      *storage.TrainerRepo
	gopherRepo       *storage.GopherRepo
	partyRepo        *storage.PartyRepo
	battleRepo       *storage.BattleRepo
	generator        *gopherkon.Generator
	evolutionService *EvolutionService
	assetsPath       string
}

func NewService(
	trainerRepo *storage.TrainerRepo,
	gopherRepo *storage.GopherRepo,
	partyRepo *storage.PartyRepo,
	battleRepo *storage.BattleRepo,
	generator *gopherkon.Generator,
	evolutionService *EvolutionService,
	assetsPath string,
) *Service {
	return &Service{
		trainerRepo:      trainerRepo,
		gopherRepo:       gopherRepo,
		partyRepo:        partyRepo,
		battleRepo:       battleRepo,
		generator:        generator,
		evolutionService: evolutionService,
		assetsPath:       assetsPath,
	}
}

// GenerateStarterGophers creates 3 starter gophers for a new trainer
func (s *Service) GenerateStarterGophers() ([]*storage.Gopher, error) {
	archetypes := []Archetype{ArchetypeHacker, ArchetypeTank, ArchetypeSpeedy, ArchetypeSupport, ArchetypeMage}
	var starters []*storage.Gopher

	for i := 0; i < 3; i++ {
		// Random archetype
		archetype := archetypes[rand.Intn(len(archetypes))]
		
		// Complexity 2-3 for starters (COMMON/UNCOMMON)
		complexity := 2 + rand.Intn(2)
		rarity := ComplexityToRarity(complexity).String()

		// Generate sprite
		result, err := s.generator.Generate(gopherkon.GenerateOptions{
			Complexity:   complexity,
			TargetRarity: rarity,
			Seed:        time.Now().UnixNano() + int64(i),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to generate sprite: %w", err)
		}

		// Save sprite
		spritePath := filepath.Join("assets", "generated", fmt.Sprintf("starter_%d_%d.png", time.Now().Unix(), i))
		if err := s.generator.SaveImage(result.Image, spritePath); err != nil {
			return nil, fmt.Errorf("failed to save sprite: %w", err)
		}

		// Generate stats
		hp, attack, defense, speed := GenerateBaseStats(archetype, rarity, 1)

		// Create gopher
		gopher := &storage.Gopher{
			Name:             GenerateGopherName(archetype),
			Level:            1,
			XP:               0,
			CurrentHP:        hp,
			MaxHP:            hp,
			Attack:           attack,
			Defense:          defense,
			Speed:            speed,
			Rarity:           rarity,
			ComplexityScore:  result.Complexity,
			SpeciesArchetype: string(archetype),
			EvolutionStage:   0,
			SpritePath:       spritePath,
			GopherkonLayers:  result.Layers,
			IsInParty:        false,
		}

		starters = append(starters, gopher)
	}

	return starters, nil
}

// GenerateStarterCard creates a card image with all 3 starter gophers
func (s *Service) GenerateStarterCard(starters []*storage.Gopher) (string, error) {
	if len(starters) != 3 {
		return "", fmt.Errorf("need exactly 3 starters for card")
	}

	// Get sprite paths
	spritePaths := make([]string, 3)
	for i, starter := range starters {
		spritePaths[i] = starter.SpritePath
	}

	// Generate card
	cardPath := filepath.Join("assets", "generated", fmt.Sprintf("starter_card_%d.png", time.Now().Unix()))
	if err := s.generator.GenerateStarterCard(spritePaths, cardPath); err != nil {
		return "", fmt.Errorf("failed to generate starter card: %w", err)
	}

	return cardPath, nil
}

// GenerateBattleCard creates a battle card with enemy on top and player on bottom
func (s *Service) GenerateBattleCard(enemyGopher, playerGopher *storage.Gopher) (string, error) {
	if enemyGopher.SpritePath == "" || playerGopher.SpritePath == "" {
		return "", fmt.Errorf("gophers must have sprite paths")
	}

	cardPath := filepath.Join("assets", "generated", fmt.Sprintf("battle_card_%s_%s.png", enemyGopher.ID[:8], playerGopher.ID[:8]))
	if err := s.generator.GenerateBattleCard(enemyGopher.SpritePath, playerGopher.SpritePath, cardPath); err != nil {
		return "", fmt.Errorf("failed to generate battle card: %w", err)
	}

	return cardPath, nil
}

// GenerateGopherCard creates a card with N gophers arranged in a grid
func (s *Service) GenerateGopherCard(gophers []*storage.Gopher, cols int) (string, error) {
	if len(gophers) == 0 {
		return "", fmt.Errorf("need at least 1 gopher")
	}

	// Get sprite paths
	spritePaths := make([]string, len(gophers))
	for i, gopher := range gophers {
		if gopher.SpritePath == "" {
			return "", fmt.Errorf("gopher %d has no sprite path", i+1)
		}
		spritePaths[i] = gopher.SpritePath
	}

	// Generate card
	cardPath := filepath.Join("assets", "generated", fmt.Sprintf("gopher_card_%d_%d.png", len(gophers), time.Now().Unix()))
	if err := s.generator.GenerateGopherCard(spritePaths, cardPath, cols); err != nil {
		return "", fmt.Errorf("failed to generate gopher card: %w", err)
	}

	return cardPath, nil
}

// GenerateWildGopher creates a wild gopher for encounters
func (s *Service) GenerateWildGopher() (*storage.Gopher, error) {
	// Determine rarity from distribution
	rarity := GetWildRarityDistribution(rand.Float64())
	minComplexity, maxComplexity := RarityToComplexityRange(rarity)
	complexity := minComplexity + rand.Intn(maxComplexity-minComplexity+1)

	// Random archetype
	archetypes := []Archetype{ArchetypeHacker, ArchetypeTank, ArchetypeSpeedy, ArchetypeSupport, ArchetypeMage}
	archetype := archetypes[rand.Intn(len(archetypes))]

	// Random level (1-10 for now)
	level := 1 + rand.Intn(10)

	// Generate sprite
	result, err := s.generator.Generate(gopherkon.GenerateOptions{
		Complexity:   complexity,
		TargetRarity: rarity.String(),
		Seed:        time.Now().UnixNano(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate sprite: %w", err)
	}

	// Save sprite
	spritePath := filepath.Join("assets", "generated", fmt.Sprintf("wild_%d.png", time.Now().UnixNano()))
	if err := s.generator.SaveImage(result.Image, spritePath); err != nil {
		return nil, fmt.Errorf("failed to save sprite: %w", err)
	}

	// Generate stats
	hp, attack, defense, speed := GenerateBaseStats(archetype, rarity.String(), level)

	// Create gopher
	gopher := &storage.Gopher{
		Name:             GenerateGopherName(archetype),
		Level:            level,
		XP:               0,
		CurrentHP:        hp,
		MaxHP:            hp,
		Attack:           attack,
		Defense:          defense,
		Speed:            speed,
		Rarity:           rarity.String(),
		ComplexityScore:  result.Complexity,
		SpeciesArchetype: string(archetype),
		EvolutionStage:   0,
		SpritePath:       spritePath,
		GopherkonLayers:  result.Layers,
		IsInParty:        false,
	}

	return gopher, nil
}

// CreateGopherWithAbilities creates a gopher and assigns abilities
func (s *Service) CreateGopherWithAbilities(gopher *storage.Gopher) error {
	// Create gopher in DB
	_, err := s.gopherRepo.Create(gopher)
	if err != nil {
		return err
	}

	// Get ability templates for archetype
	abilityTemplates := GetAbilitiesForArchetype(Archetype(gopher.SpeciesArchetype))
	
	// Assign first 2 abilities (more unlock at higher levels)
	numAbilities := 2
	if gopher.Level >= 10 {
		numAbilities = 3
	}
	if gopher.Level >= 20 {
		numAbilities = 4
	}

	if numAbilities > len(abilityTemplates) {
		numAbilities = len(abilityTemplates)
	}

	// Create abilities (we'll store them in memory for now, can persist later)
	// For now, abilities are stored in the Gopher struct's Abilities field
	// which we'll need to handle differently since DB stores them separately

	return nil
}

// StorageGopherToGameGopher converts storage.Gopher to game.Gopher with abilities
func (s *Service) StorageGopherToGameGopher(storageGopher *storage.Gopher) (*Gopher, error) {
	gameGopher := &Gopher{
		ID:               storageGopher.ID,
		TrainerID:       storageGopher.TrainerID,
		Name:             storageGopher.Name,
		Level:            storageGopher.Level,
		XP:               storageGopher.XP,
		CurrentHP:        storageGopher.CurrentHP,
		MaxHP:            storageGopher.MaxHP,
		Attack:           storageGopher.Attack,
		Defense:          storageGopher.Defense,
		Speed:            storageGopher.Speed,
		Rarity:           storageGopher.Rarity,
		ComplexityScore:  storageGopher.ComplexityScore,
		SpeciesArchetype: storageGopher.SpeciesArchetype,
		EvolutionStage:   storageGopher.EvolutionStage,
		SpritePath:       storageGopher.SpritePath,
		GopherkonLayers:  storageGopher.GopherkonLayers,
		IsInParty:        storageGopher.IsInParty,
		PCSlot:           storageGopher.PCSlot,
		Abilities:        []*Ability{},
	}

	// Create abilities for this gopher
	abilityTemplates := GetAbilitiesForArchetype(Archetype(gameGopher.SpeciesArchetype))
	numAbilities := 2
	if gameGopher.Level >= 10 {
		numAbilities = 3
	}
	if gameGopher.Level >= 20 {
		numAbilities = 4
	}

	if numAbilities > len(abilityTemplates) {
		numAbilities = len(abilityTemplates)
	}

	for i := 0; i < numAbilities; i++ {
		ability, err := CreateAbilityFromTemplate(abilityTemplates[i], fmt.Sprintf("%s_ability_%d", gameGopher.ID, i))
		if err != nil {
			continue
		}
		gameGopher.Abilities = append(gameGopher.Abilities, ability)
	}

	return gameGopher, nil
}

// CheckEvolution checks if a gopher should evolve
func (s *Service) CheckEvolution(gameGopher *Gopher) (evolved bool, message string) {
	return s.evolutionService.CheckAndEvolve(gameGopher)
}

