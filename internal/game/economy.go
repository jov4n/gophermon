package game

import (
	"fmt"
)

// Item types
const (
	ItemTypePotion        = "POTION"
	ItemTypeRevive       = "REVIVE"
	ItemTypeXPBooster    = "XP_BOOSTER"
	ItemTypeEvolutionStone = "EVOLUTION_STONE"
	ItemTypeShinyCharm   = "SHINY_CHARM"
)

// Item prices
const (
	PricePotion        = 50
	PriceRevive        = 100
	PriceXPBooster     = 200
	PriceEvolutionStone = 500
	PriceShinyCharm    = 1000
)

// Item effects
const (
	PotionHealAmount    = 50 // Heals 50 HP
	ReviveHealAmount    = 1  // Restores to 1 HP
	XPBoosterMultiplier = 1.5 // 1.5x XP for next battle
	ShinyCharmRate      = 1.0 / 2048.0 // Doubles shiny rate
)

type ItemService struct {
	trainerRepo TrainerRepoInterface
	itemRepo    ItemRepoInterface
}

func NewItemService(trainerRepo TrainerRepoInterface, itemRepo ItemRepoInterface) *ItemService {
	return &ItemService{
		trainerRepo: trainerRepo,
		itemRepo:    itemRepo,
	}
}

func (s *ItemService) GetItemPrice(itemType string) int {
	switch itemType {
	case ItemTypePotion:
		return PricePotion
	case ItemTypeRevive:
		return PriceRevive
	case ItemTypeXPBooster:
		return PriceXPBooster
	case ItemTypeEvolutionStone:
		return PriceEvolutionStone
	case ItemTypeShinyCharm:
		return PriceShinyCharm
	default:
		return 0
	}
}

func (s *ItemService) BuyItem(trainerID, itemType string, quantity int) error {
	price := s.GetItemPrice(itemType)
	if price == 0 {
		return fmt.Errorf("invalid item type")
	}

	totalCost := price * quantity
	if err := s.trainerRepo.RemoveCurrency(trainerID, totalCost); err != nil {
		return fmt.Errorf("insufficient currency: %w", err)
	}

	if err := s.itemRepo.AddItem(trainerID, itemType, quantity); err != nil {
		// Refund on error
		s.trainerRepo.AddCurrency(trainerID, totalCost)
		return fmt.Errorf("failed to add item: %w", err)
	}

	return nil
}

func (s *ItemService) UsePotion(gopher *Gopher) error {
	healAmount := PotionHealAmount
	newHP := gopher.CurrentHP + healAmount
	if newHP > gopher.MaxHP {
		newHP = gopher.MaxHP
	}
	gopher.CurrentHP = newHP
	return nil
}

func (s *ItemService) UseRevive(gopher *Gopher) error {
	if gopher.CurrentHP > 0 {
		return fmt.Errorf("gopher is not fainted")
	}
	gopher.CurrentHP = ReviveHealAmount
	return nil
}

func (s *ItemService) GetShinyRateMultiplier(trainerID string) float64 {
	// Check if trainer has shiny charm
	if qty, err := s.itemRepo.GetItemQuantity(trainerID, ItemTypeShinyCharm); err == nil && qty > 0 {
		return 2.0 // Doubles shiny rate
	}
	return 1.0
}

