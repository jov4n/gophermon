package game

import (
	"fmt"
	"math/rand"
	"strings"
)

// BattleState represents the current state of a battle
type BattleState struct {
	ID           string
	ChannelID    string
	MessageID    string
	TrainerID    string
	OpponentType string
	PlayerGopher *Gopher
	EnemyGopher  *Gopher
	TurnOwner    string // "PLAYER" or "ENEMY"
	State        string // "ACTIVE", "WON", "LOST", "ESCAPED"
	Log          []string
}

// NewBattleState creates a new battle state
func NewBattleState(trainerID, channelID string, playerGopher, enemyGopher *Gopher) *BattleState {
	return &BattleState{
		TrainerID:    trainerID,
		ChannelID:    channelID,
		PlayerGopher: playerGopher,
		EnemyGopher:  enemyGopher,
		TurnOwner:    "PLAYER",
		State:        "ACTIVE",
		Log:          []string{fmt.Sprintf("A wild %s appeared!", enemyGopher.Name)},
	}
}

// PlayerAction executes a player action
func (bs *BattleState) PlayerAction(action string, abilityIndex int) ([]string, error) {
	if bs.State != "ACTIVE" {
		return []string{"Battle is already over!"}, nil
	}

	if bs.TurnOwner != "PLAYER" {
		return []string{"It's not your turn!"}, nil
	}

	messages := []string{}

	switch action {
	case "fight":
		if abilityIndex < 0 || abilityIndex >= len(bs.PlayerGopher.Abilities) {
			return []string{"Invalid ability!"}, nil
		}

		ability := bs.PlayerGopher.Abilities[abilityIndex]
		msgs, err := ability.EffectFunc(bs, bs.PlayerGopher, bs.EnemyGopher)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msgs...)

		// Check if enemy is defeated
		if bs.EnemyGopher.CurrentHP <= 0 {
			bs.State = "WON"
			xpGain := bs.calculateXPGain()
			leveledUp, newLevel := bs.PlayerGopher.AddXP(xpGain)
			messages = append(messages, fmt.Sprintf("%s was defeated!", bs.EnemyGopher.Name))
			messages = append(messages, fmt.Sprintf("%s gained %d XP!", bs.PlayerGopher.Name, xpGain))
			if leveledUp {
				messages = append(messages, fmt.Sprintf("%s leveled up to level %d!", bs.PlayerGopher.Name, newLevel))
			}
			return messages, nil
		}

		// Enemy turn
		bs.TurnOwner = "ENEMY"
		enemyMsgs := bs.enemyTurn()
		messages = append(messages, enemyMsgs...)

		if bs.PlayerGopher.CurrentHP <= 0 {
			bs.State = "LOST"
			messages = append(messages, fmt.Sprintf("%s was defeated!", bs.PlayerGopher.Name))
		}

	case "run":
		// 70% chance to escape
		if rand.Float64() < 0.7 {
			bs.State = "ESCAPED"
			messages = append(messages, "Got away safely!")
		} else {
			messages = append(messages, "Couldn't escape!")
			// Enemy turn
			bs.TurnOwner = "ENEMY"
			enemyMsgs := bs.enemyTurn()
			messages = append(messages, enemyMsgs...)
		}

	case "throw_net":
		captureChance := bs.calculateCaptureChance()
		if rand.Float64() < captureChance {
			bs.State = "WON"
			messages = append(messages, fmt.Sprintf("Successfully captured %s!", bs.EnemyGopher.Name))
		} else {
			messages = append(messages, "The gopher broke free!")
			// Enemy turn
			bs.TurnOwner = "ENEMY"
			enemyMsgs := bs.enemyTurn()
			messages = append(messages, enemyMsgs...)
		}

	default:
		return []string{"Unknown action!"}, nil
	}

	bs.Log = append(bs.Log, messages...)
	return messages, nil
}

// enemyTurn executes the enemy's turn
func (bs *BattleState) enemyTurn() []string {
	if len(bs.EnemyGopher.Abilities) == 0 {
		return []string{fmt.Sprintf("%s has no abilities!", bs.EnemyGopher.Name)}
	}

	// Simple AI: random ability
	abilityIndex := rand.Intn(len(bs.EnemyGopher.Abilities))
	ability := bs.EnemyGopher.Abilities[abilityIndex]

	messages := []string{}
	msgs, err := ability.EffectFunc(bs, bs.EnemyGopher, bs.PlayerGopher)
	if err == nil {
		messages = append(messages, msgs...)
	}

	bs.TurnOwner = "PLAYER"
	return messages
}

// calculateXPGain calculates XP gained from defeating enemy
func (bs *BattleState) calculateXPGain() int {
	baseXP := bs.EnemyGopher.Level * 10
	
	// Rarity bonus
	rarityMultiplier := 1.0
	switch bs.EnemyGopher.Rarity {
	case "COMMON":
		rarityMultiplier = 1.0
	case "UNCOMMON":
		rarityMultiplier = 1.2
	case "RARE":
		rarityMultiplier = 1.5
	case "EPIC":
		rarityMultiplier = 2.0
	case "LEGENDARY":
		rarityMultiplier = 3.0
	}
	
	return int(float64(baseXP) * rarityMultiplier)
}

// calculateCaptureChance calculates the chance to capture a gopher
func (bs *BattleState) calculateCaptureChance() float64 {
	// Base chance based on HP remaining
	hpPercent := float64(bs.EnemyGopher.CurrentHP) / float64(bs.EnemyGopher.MaxHP)
	baseChance := 1.0 - hpPercent // Lower HP = higher chance
	
	// Rarity penalty
	rarityPenalty := 0.0
	switch bs.EnemyGopher.Rarity {
	case "COMMON":
		rarityPenalty = 0.0
	case "UNCOMMON":
		rarityPenalty = 0.1
	case "RARE":
		rarityPenalty = 0.2
	case "EPIC":
		rarityPenalty = 0.3
	case "LEGENDARY":
		rarityPenalty = 0.4
	}
	
	chance := baseChance * (1.0 - rarityPenalty)
	
	// Minimum 5% chance, maximum 90% chance
	if chance < 0.05 {
		chance = 0.05
	}
	if chance > 0.90 {
		chance = 0.90
	}
	
	return chance
}

// GetHPBar creates a visual HP bar
func GetHPBar(current, max int, length int) string {
	if max == 0 {
		return strings.Repeat("░", length)
	}
	
	filled := int(float64(current) / float64(max) * float64(length))
	if filled > length {
		filled = length
	}
	if filled < 0 {
		filled = 0
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", length-filled)
	return fmt.Sprintf("%s %d/%d", bar, current, max)
}

