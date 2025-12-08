package game

import (
	"fmt"
	"image"
	"math/rand"
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

		// Encode sprite to base64
		spriteData, err := s.generator.EncodeImageToBase64(result.Image)
		if err != nil {
			return nil, fmt.Errorf("failed to encode sprite: %w", err)
		}

		// Generate stats
		hp, attack, defense, speed := GenerateBaseStats(archetype, rarity, 1)

		// Assign types: primary type from archetype, chance for secondary type
		primaryType := GetTypeFromArchetype(archetype)
		secondaryType := ""
		// 30% chance for dual type (rarer gophers more likely)
		if rand.Float64() < 0.3 || rarity == "RARE" || rarity == "EPIC" || rarity == "LEGENDARY" {
			secondaryType = string(GetRandomSecondaryType(primaryType))
		}

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
			PrimaryType:      string(primaryType),
			SecondaryType:    secondaryType,
			SpritePath:       "", // No longer used
			SpriteData:       spriteData,
			GopherkonLayers:  result.Layers,
			IsInParty:        false,
		}

		starters = append(starters, gopher)
	}

	return starters, nil
}

// GenerateStarterCard creates a card image with all 3 starter gophers and returns base64
func (s *Service) GenerateStarterCard(starters []*storage.Gopher) (string, error) {
	if len(starters) != 3 {
		return "", fmt.Errorf("need exactly 3 starters for card")
	}

	// Get sprite data (base64) and decode to images
	spriteImages := make([]image.Image, 3)
	for i, starter := range starters {
		if starter.SpriteData != "" {
			img, err := s.generator.DecodeImageFromBase64(starter.SpriteData)
			if err == nil {
				spriteImages[i] = img
			}
		} else if starter.SpritePath != "" {
			// Fallback for old gophers with file paths
			img, err := s.generator.LoadImageFromPath(starter.SpritePath)
			if err == nil {
				spriteImages[i] = img
			}
		}
	}

	// Generate card and return base64
	cardBase64, err := s.generator.GenerateStarterCardFromImagesToBase64(spriteImages)
	if err != nil {
		return "", fmt.Errorf("failed to generate starter card: %w", err)
	}

	return cardBase64, nil
}

// GenerateBattleCard creates a battle card with enemy on top and player on bottom and returns base64
func (s *Service) GenerateBattleCard(enemyGopher, playerGopher *storage.Gopher) (string, error) {
	// Load enemy sprite
	var enemyImg image.Image
	var err error
	if enemyGopher.SpriteData != "" {
		enemyImg, err = s.generator.DecodeImageFromBase64(enemyGopher.SpriteData)
	} else if enemyGopher.SpritePath != "" {
		enemyImg, err = s.generator.LoadImageFromPath(enemyGopher.SpritePath)
	} else {
		return "", fmt.Errorf("enemy gopher has no sprite data")
	}
	if err != nil {
		return "", fmt.Errorf("failed to load enemy sprite: %w", err)
	}

	// Load player sprite
	var playerImg image.Image
	if playerGopher.SpriteData != "" {
		playerImg, err = s.generator.DecodeImageFromBase64(playerGopher.SpriteData)
	} else if playerGopher.SpritePath != "" {
		playerImg, err = s.generator.LoadImageFromPath(playerGopher.SpritePath)
	} else {
		return "", fmt.Errorf("player gopher has no sprite data")
	}
	if err != nil {
		return "", fmt.Errorf("failed to load player sprite: %w", err)
	}

	// Generate card and return base64 with names and levels
	cardBase64, err := s.generator.GenerateBattleCardFromImagesToBase64(
		enemyImg, playerImg, 
		enemyGopher.Name, playerGopher.Name,
		enemyGopher.Level, playerGopher.Level,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate battle card: %w", err)
	}

	return cardBase64, nil
}

// GenerateGopherCard creates a card with N gophers arranged in a grid and returns base64
func (s *Service) GenerateGopherCard(gophers []*storage.Gopher, cols int) (string, error) {
	if len(gophers) == 0 {
		return "", fmt.Errorf("need at least 1 gopher")
	}

	// Get sprite images from base64 or file paths
	spriteImages := make([]image.Image, len(gophers))
	for i, gopher := range gophers {
		var img image.Image
		var err error
		if gopher.SpriteData != "" {
			img, err = s.generator.DecodeImageFromBase64(gopher.SpriteData)
		} else if gopher.SpritePath != "" {
			img, err = s.generator.LoadImageFromPath(gopher.SpritePath)
		} else {
			return "", fmt.Errorf("gopher %d has no sprite data", i+1)
		}
		if err != nil {
			return "", fmt.Errorf("failed to load gopher %d sprite: %w", i+1, err)
		}
		spriteImages[i] = img
	}

	// Generate card and return base64
	cardBase64, err := s.generator.GenerateGopherCardFromImagesToBase64(spriteImages, cols)
	if err != nil {
		return "", fmt.Errorf("failed to generate gopher card: %w", err)
	}

	return cardBase64, nil
}

// GenerateGopherWithRarity creates a gopher with a specific rarity
func (s *Service) GenerateGopherWithRarity(targetRarity Rarity, seedOffset int64) (*storage.Gopher, error) {
	// Random archetype
	archetypes := []Archetype{ArchetypeHacker, ArchetypeTank, ArchetypeSpeedy, ArchetypeSupport, ArchetypeMage}
	archetype := archetypes[rand.Intn(len(archetypes))]

	// Random level (1-10 for now)
	level := 1 + rand.Intn(10)

	// Generate sprite with specific rarity
	result, err := s.generator.Generate(gopherkon.GenerateOptions{
		TargetRarity: targetRarity.String(),
		Seed:        time.Now().UnixNano() + seedOffset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate sprite: %w", err)
	}

	// Encode sprite to base64
	spriteData, err := s.generator.EncodeImageToBase64(result.Image)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sprite: %w", err)
	}

		// Generate stats
		hp, attack, defense, speed := GenerateBaseStats(archetype, targetRarity.String(), level)

		// Assign types: primary type from archetype, chance for secondary type
		primaryType := GetTypeFromArchetype(archetype)
		secondaryType := ""
		// 30% chance for dual type (rarer gophers more likely)
		if rand.Float64() < 0.3 || targetRarity == RarityRare || targetRarity == RarityEpic || targetRarity == RarityLegendary {
			secondaryType = string(GetRandomSecondaryType(primaryType))
		}

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
			Rarity:           targetRarity.String(),
			ComplexityScore:  result.Complexity,
			SpeciesArchetype: string(archetype),
			EvolutionStage:   0,
			PrimaryType:      string(primaryType),
			SecondaryType:    secondaryType,
			SpritePath:       "", // No longer used
			SpriteData:       spriteData,
			GopherkonLayers:  result.Layers,
			IsInParty:        false,
		}

	return gopher, nil
}

// GenerateWildGopher creates a wild gopher for encounters
func (s *Service) GenerateWildGopher() (*storage.Gopher, error) {
	// Determine rarity from distribution
	rarity := GetWildRarityDistribution(rand.Float64())
	return s.GenerateGopherWithRarity(rarity, 0)
}

// CreateGopherWithAbilities creates a gopher and assigns abilities
func (s *Service) CreateGopherWithAbilities(gopher *storage.Gopher) error {
	// Create gopher in DB
	_, err := s.gopherRepo.Create(gopher)
	if err != nil {
		return err
	}

	// Get ability templates for archetype
	abilityTemplates := GetAbilitiesForArchetype(Archetype(gopher.SpeciesArchetype), gopher.EvolutionStage, gopher.Rarity)
	
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
	// Convert types (with fallback to archetype if not set)
	primaryType := GopherType(storageGopher.PrimaryType)
	if primaryType == "" {
		primaryType = GetTypeFromArchetype(Archetype(storageGopher.SpeciesArchetype))
	}
	secondaryType := GopherType(storageGopher.SecondaryType)

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
		PrimaryType:      primaryType,
		SecondaryType:    secondaryType,
		SpritePath:       storageGopher.SpritePath,
		SpriteData:       storageGopher.SpriteData,
		GopherkonLayers:  storageGopher.GopherkonLayers,
		IsInParty:        storageGopher.IsInParty,
		PCSlot:           storageGopher.PCSlot,
		Abilities:        []*Ability{},
		StatusEffects:    []*StatusEffect{},
		BaseAttack:       storageGopher.Attack,
		BaseDefense:      storageGopher.Defense,
		BaseSpeed:        storageGopher.Speed,
	}

	// Create abilities for this gopher
	abilityTemplates := GetAbilitiesForArchetype(Archetype(gameGopher.SpeciesArchetype), gameGopher.EvolutionStage, gameGopher.Rarity)
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

// CheckAndHandleBlackout checks if all party members are dead and handles blackout
// Returns true if blackout occurred, along with a message
func (s *Service) CheckAndHandleBlackout(trainerID string) (blackedOut bool, message string) {
	// Get all party members
	party, err := s.gopherRepo.GetParty(trainerID)
	if err != nil {
		return false, ""
	}

	if len(party) == 0 {
		return false, "" // No party, can't black out
	}

	// Check if all party members are dead (HP <= 0)
	allDead := true
	for _, gopher := range party {
		if gopher.CurrentHP > 0 {
			allDead = false
			break
		}
	}

	if !allDead {
		return false, ""
	}

	// All party members are dead - trigger blackout
	// Restore all party members to full health
	for _, gopher := range party {
		gopher.CurrentHP = gopher.MaxHP
		if err := s.gopherRepo.Update(gopher); err != nil {
			// Log error but continue with other gophers
			continue
		}
	}

	message = "💀 **You blacked out!**\n"
	message += "All your gophers fainted!\n"
	message += "You were taken to a safe place and your party was restored to full health."

	return true, message
}

