package game

import (
	"fmt"
	"math/rand"
	"strings"
)

// BattleState represents the current state of a battle
type BattleState struct {
	ID                string
	ChannelID         string
	MessageID         string
	TrainerID         string
	OpponentType      string
	PlayerGopher      *Gopher
	EnemyGopher       *Gopher
	PlayerParty       []*Gopher  // All party members (for swapping and XP distribution)
	ParticipatingGophers []*Gopher // Gophers that have participated in battle (for XP)
	TurnOwner         string      // "PLAYER" or "ENEMY"
	State             string      // "ACTIVE", "WON", "LOST", "ESCAPED"
	Log               []string
	EventManager      *EventManager // Event manager for event bonuses
}

// NewBattleState creates a new battle state
func NewBattleState(trainerID, channelID string, playerGopher, enemyGopher *Gopher, playerParty []*Gopher, eventManager *EventManager) *BattleState {
	// Initialize participating gophers with the active gopher
	participating := []*Gopher{playerGopher}
	
	// Apply stat boost event to all gophers in battle
	if eventManager != nil {
		statBoost := eventManager.GetStatBoostMultiplier()
		if statBoost > 1.0 {
			// Ensure base stats are initialized, then apply boost
			applyStatBoost := func(gopher *Gopher) {
				// Initialize base stats if needed
				if gopher.BaseAttack == 0 {
					gopher.BaseAttack = gopher.Attack
				}
				if gopher.BaseDefense == 0 {
					gopher.BaseDefense = gopher.Defense
				}
				if gopher.BaseSpeed == 0 {
					gopher.BaseSpeed = gopher.Speed
				}
				// Recalculate stats first (applies status effects)
				gopher.RecalculateStats()
				// Then apply event stat boost on top (multiply current stats)
				gopher.Attack = int(float64(gopher.Attack) * statBoost)
				gopher.Defense = int(float64(gopher.Defense) * statBoost)
				gopher.Speed = int(float64(gopher.Speed) * statBoost)
			}
			
			applyStatBoost(playerGopher)
			applyStatBoost(enemyGopher)
			for _, gopher := range playerParty {
				if gopher.ID != playerGopher.ID {
					applyStatBoost(gopher)
				}
			}
		}
	}
	
	return &BattleState{
		TrainerID:          trainerID,
		ChannelID:          channelID,
		PlayerGopher:       playerGopher,
		EnemyGopher:        enemyGopher,
		PlayerParty:        playerParty,
		ParticipatingGophers: participating,
		TurnOwner:          "PLAYER",
		State:              "ACTIVE",
		Log:                []string{fmt.Sprintf("A wild %s appeared!", enemyGopher.Name)},
		EventManager:       eventManager,
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
	
	// Process status effects at start of turn
	statusMsgs := bs.PlayerGopher.ProcessStatusEffects()
	messages = append(messages, statusMsgs...)
	
	// Check if player is asleep or paralyzed
	if bs.PlayerGopher.HasStatusEffect(StatusSleep) {
		// Check if sleep wears off
		for _, effect := range bs.PlayerGopher.StatusEffects {
			if effect.Type == StatusSleep {
				if rand.Float64() < 0.3 { // 30% chance to wake up
					bs.PlayerGopher.RemoveStatusEffect(StatusSleep)
					messages = append(messages, fmt.Sprintf("%s woke up!", bs.PlayerGopher.Name))
				} else {
					messages = append(messages, fmt.Sprintf("%s is fast asleep!", bs.PlayerGopher.Name))
					bs.TurnOwner = "ENEMY"
					enemyMsgs := bs.enemyTurn()
					messages = append(messages, enemyMsgs...)
					bs.Log = append(bs.Log, messages...)
					return messages, nil
				}
			}
		}
	}
	
	// Check paralysis
	if bs.PlayerGopher.HasStatusEffect(StatusParalysis) {
		if rand.Float64() < 0.25 { // 25% chance to be paralyzed
			messages = append(messages, fmt.Sprintf("%s is paralyzed! It can't move!", bs.PlayerGopher.Name))
			bs.TurnOwner = "ENEMY"
			enemyMsgs := bs.enemyTurn()
			messages = append(messages, enemyMsgs...)
			bs.Log = append(bs.Log, messages...)
			return messages, nil
		}
	}
	
	// Check confusion
	confused := bs.PlayerGopher.HasStatusEffect(StatusConfusion)
	if confused && action == "fight" {
		if rand.Float64() < 0.33 { // 33% chance to hurt self
			damage := bs.PlayerGopher.MaxHP / 8
			bs.PlayerGopher.CurrentHP -= damage
			if bs.PlayerGopher.CurrentHP < 0 {
				bs.PlayerGopher.CurrentHP = 0
			}
			messages = append(messages, fmt.Sprintf("%s is confused! It hurt itself in confusion for %d damage!", bs.PlayerGopher.Name, damage))
			bs.TurnOwner = "ENEMY"
			enemyMsgs := bs.enemyTurn()
			messages = append(messages, enemyMsgs...)
			if bs.PlayerGopher.CurrentHP <= 0 {
				bs.State = "LOST"
				messages = append(messages, fmt.Sprintf("%s was defeated!", bs.PlayerGopher.Name))
			}
			bs.Log = append(bs.Log, messages...)
			return messages, nil
		}
	}
	
	// Check protect status
	if bs.PlayerGopher.HasStatusEffect(StatusProtect) && action == "fight" {
		// Protection is consumed when used, so remove it
		bs.PlayerGopher.RemoveStatusEffect(StatusProtect)
	}

	switch action {
	case "fight":
		if abilityIndex < 0 || abilityIndex >= len(bs.PlayerGopher.Abilities) {
			return []string{"Invalid ability!"}, nil
		}

		ability := bs.PlayerGopher.Abilities[abilityIndex]
		
		// Check if enemy has protect status
		if bs.EnemyGopher.HasStatusEffect(StatusProtect) {
			bs.EnemyGopher.RemoveStatusEffect(StatusProtect)
			messages = append(messages, fmt.Sprintf("%s was protected from the attack!", bs.EnemyGopher.Name))
		} else {
			msgs, err := ability.EffectFunc(bs, bs.PlayerGopher, bs.EnemyGopher)
			if err != nil {
				return nil, err
			}
			messages = append(messages, msgs...)
		}

		// Check if enemy is defeated
		if bs.EnemyGopher.CurrentHP <= 0 {
			bs.State = "WON"
			xpGain := bs.calculateXPGain()
			
			// Give XP to all participating gophers (including fainted ones)
			for _, gopher := range bs.ParticipatingGophers {
				leveledUp, newLevel := gopher.AddXP(xpGain)
				xpBar := GetXPBar(gopher.XP, gopher.Level, 10)
				messages = append(messages, fmt.Sprintf("%s gained %d XP! %s", gopher.Name, xpGain, xpBar))
				if leveledUp {
					messages = append(messages, fmt.Sprintf("%s leveled up to level %d! 🎉", gopher.Name, newLevel))
				}
			}
			
			messages = append(messages, fmt.Sprintf("%s was defeated!", bs.EnemyGopher.Name))
			return messages, nil
		}

		// Enemy turn
		bs.TurnOwner = "ENEMY"
		enemyMsgs := bs.enemyTurn()
		messages = append(messages, enemyMsgs...)

		// Check if player gopher fainted and try auto-swap
		if bs.PlayerGopher.CurrentHP <= 0 {
			swapped, swapMsgs := bs.tryAutoSwap()
			messages = append(messages, swapMsgs...)
			if !swapped {
				// Battle lost if no swap occurred
				bs.State = "LOST"
			}
		}

	case "swap":
		// Party swap - abilityIndex contains the party member index to swap to
		if abilityIndex < 0 || abilityIndex >= len(bs.PlayerParty) {
			return []string{"Invalid party member!"}, nil
		}
		
		newGopher := bs.PlayerParty[abilityIndex]
		if newGopher.CurrentHP <= 0 {
			return []string{fmt.Sprintf("%s is fainted and can't battle!", newGopher.Name)}, nil
		}
		
		if newGopher.ID == bs.PlayerGopher.ID {
			return []string{"That gopher is already in battle!"}, nil
		}
		
		// Swap gophers
		oldGopher := bs.PlayerGopher
		bs.PlayerGopher = newGopher
		
		// Add new gopher to participating list if not already there
		found := false
		for _, g := range bs.ParticipatingGophers {
			if g.ID == newGopher.ID {
				found = true
				break
			}
		}
		if !found {
			bs.ParticipatingGophers = append(bs.ParticipatingGophers, newGopher)
		}
		
		messages = append(messages, fmt.Sprintf("%s, come back!", oldGopher.Name))
		messages = append(messages, fmt.Sprintf("Go, %s!", newGopher.Name))
		
		// Enemy gets a free turn after swap (like Pokemon)
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
			xpGain := bs.calculateXPGain()
			
			// Give XP to all participating gophers on capture (including fainted ones)
			for _, gopher := range bs.ParticipatingGophers {
				leveledUp, newLevel := gopher.AddXP(xpGain)
				xpBar := GetXPBar(gopher.XP, gopher.Level, 10)
				messages = append(messages, fmt.Sprintf("%s gained %d XP! %s", gopher.Name, xpGain, xpBar))
				if leveledUp {
					messages = append(messages, fmt.Sprintf("%s leveled up to level %d! 🎉", gopher.Name, newLevel))
				}
			}
			
			messages = append(messages, fmt.Sprintf("Successfully captured %s!", bs.EnemyGopher.Name))
		} else {
			messages = append(messages, "The gopher broke free!")
			// Enemy turn
			bs.TurnOwner = "ENEMY"
			enemyMsgs := bs.enemyTurn()
			messages = append(messages, enemyMsgs...)
			
			// Check if player gopher fainted after enemy turn and try auto-swap
			if bs.PlayerGopher.CurrentHP <= 0 {
				swapped, swapMsgs := bs.tryAutoSwap()
				messages = append(messages, swapMsgs...)
				if !swapped {
					// Battle lost if no swap occurred
					bs.State = "LOST"
				}
			}
		}

	default:
		return []string{"Unknown action!"}, nil
	}

	bs.Log = append(bs.Log, messages...)
	return messages, nil
}

// tryAutoSwap attempts to swap to another party member if current gopher fainted
// Returns true if swap occurred, false otherwise
func (bs *BattleState) tryAutoSwap() (swapped bool, messages []string) {
	if bs.PlayerGopher.CurrentHP > 0 {
		return false, nil
	}
	
	messages = append(messages, fmt.Sprintf("%s was defeated!", bs.PlayerGopher.Name))
	
	// Find next available gopher
	var nextGopher *Gopher
	for _, gopher := range bs.PlayerParty {
		if gopher.ID != bs.PlayerGopher.ID && gopher.CurrentHP > 0 {
			nextGopher = gopher
			break
		}
	}
	
	if nextGopher != nil {
		// Swap to next available gopher
		oldGopher := bs.PlayerGopher
		bs.PlayerGopher = nextGopher
		
		// Add new gopher to participating list if not already there
		found := false
		for _, g := range bs.ParticipatingGophers {
			if g.ID == nextGopher.ID {
				found = true
				break
			}
		}
		if !found {
			bs.ParticipatingGophers = append(bs.ParticipatingGophers, nextGopher)
		}
		
		messages = append(messages, fmt.Sprintf("%s, come back!", oldGopher.Name))
		messages = append(messages, fmt.Sprintf("Go, %s!", nextGopher.Name))
		bs.TurnOwner = "PLAYER" // Player gets to act with new gopher
		return true, messages
	} else {
		// No more gophers available - battle lost
		bs.State = "LOST"
		return false, messages
	}
}

// enemyTurn executes the enemy's turn
func (bs *BattleState) enemyTurn() []string {
	messages := []string{}
	
	// Process status effects at start of turn
	statusMsgs := bs.EnemyGopher.ProcessStatusEffects()
	messages = append(messages, statusMsgs...)
	
	// Check if enemy is asleep or paralyzed
	if bs.EnemyGopher.HasStatusEffect(StatusSleep) {
		// Check if sleep wears off
		for _, effect := range bs.EnemyGopher.StatusEffects {
			if effect.Type == StatusSleep {
				if rand.Float64() < 0.3 { // 30% chance to wake up
					bs.EnemyGopher.RemoveStatusEffect(StatusSleep)
					messages = append(messages, fmt.Sprintf("%s woke up!", bs.EnemyGopher.Name))
				} else {
					messages = append(messages, fmt.Sprintf("%s is fast asleep!", bs.EnemyGopher.Name))
					bs.TurnOwner = "PLAYER"
					return messages
				}
			}
		}
	}
	
	// Check paralysis
	if bs.EnemyGopher.HasStatusEffect(StatusParalysis) {
		if rand.Float64() < 0.25 { // 25% chance to be paralyzed
			messages = append(messages, fmt.Sprintf("%s is paralyzed! It can't move!", bs.EnemyGopher.Name))
			bs.TurnOwner = "PLAYER"
			return messages
		}
	}
	
	// Check confusion
	confused := bs.EnemyGopher.HasStatusEffect(StatusConfusion)
	if confused {
		if rand.Float64() < 0.33 { // 33% chance to hurt self
			damage := bs.EnemyGopher.MaxHP / 8
			bs.EnemyGopher.CurrentHP -= damage
			if bs.EnemyGopher.CurrentHP < 0 {
				bs.EnemyGopher.CurrentHP = 0
			}
			messages = append(messages, fmt.Sprintf("%s is confused! It hurt itself in confusion for %d damage!", bs.EnemyGopher.Name, damage))
			bs.TurnOwner = "PLAYER"
			return messages
		}
	}
	
	if len(bs.EnemyGopher.Abilities) == 0 {
		return []string{fmt.Sprintf("%s has no abilities!", bs.EnemyGopher.Name)}
	}

	// Simple AI: random ability
	abilityIndex := rand.Intn(len(bs.EnemyGopher.Abilities))
	ability := bs.EnemyGopher.Abilities[abilityIndex]

	// Check if player has protect status
	if bs.PlayerGopher.HasStatusEffect(StatusProtect) {
		bs.PlayerGopher.RemoveStatusEffect(StatusProtect)
		messages = append(messages, fmt.Sprintf("%s was protected from the attack!", bs.PlayerGopher.Name))
	} else {
		msgs, err := ability.EffectFunc(bs, bs.EnemyGopher, bs.PlayerGopher)
		if err == nil {
			messages = append(messages, msgs...)
		}
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
	
	// Apply event multiplier (e.g., Double XP event)
	eventMultiplier := 1.0
	if bs.EventManager != nil {
		eventMultiplier = bs.EventManager.GetXPMultiplier()
	}
	
	return int(float64(baseXP) * rarityMultiplier * eventMultiplier)
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
	
	// Apply event multiplier (e.g., Lucky Day event)
	if bs.EventManager != nil {
		chance *= bs.EventManager.GetCaptureRateMultiplier()
	}
	
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

// GetXPBar creates a visual XP bar showing progress to next level
func GetXPBar(currentXP, currentLevel int, length int) string {
	// Ensure XP is never negative
	if currentXP < 0 {
		currentXP = 0
	}
	if currentLevel < 1 {
		currentLevel = 1
	}
	
	// XPNeeded returns total cumulative XP to reach a level
	// For level 1: 50, level 2: 200, level 3: 450, level 4: 800, etc.
	// Gophers are created at any level with 0 XP, not the XP required for that level
	// The AddXP function checks: g.XP >= XPNeeded(g.Level+1) to level up
	// Examples:
	//   - Level 1 gopher with 0 XP needs 200 XP (XPNeeded(2)) total to reach level 2
	//   - Level 3 gopher with 0 XP needs 800 XP (XPNeeded(4)) total to reach level 4
	
	// Calculate XP needed from current level to next level
	xpForNextLevel := XPNeeded(currentLevel + 1)
	xpForCurrentLevel := XPNeeded(currentLevel)
	xpNeededForNextLevel := xpForNextLevel - xpForCurrentLevel
	
	if xpNeededForNextLevel <= 0 {
		return strings.Repeat("█", length) + " MAX"
	}
	
	// Calculate progress: how much XP the gopher has towards the next level
	// Since gophers start at level N with 0 XP and need xpNeededForNextLevel XP to reach level N+1,
	// we use currentXP directly as progress (clamped to valid range)
	// For example: Level 1 gopher with 48 XP needs 150 XP (200-50) to reach level 2, so progress is 48/150
	xpProgress := currentXP
	if xpProgress > xpNeededForNextLevel {
		// Gopher has more XP than needed (should have leveled up, but handle edge case)
		xpProgress = xpNeededForNextLevel
	}
	
	filled := int(float64(xpProgress) / float64(xpNeededForNextLevel) * float64(length))
	if filled > length {
		filled = length
	}
	if filled < 0 {
		filled = 0
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", length-filled)
	return fmt.Sprintf("%s %d/%d", bar, xpProgress, xpNeededForNextLevel)
}

