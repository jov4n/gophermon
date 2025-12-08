package game

import (
	"fmt"
	"math/rand"
)

// Targeting defines who an ability affects
type Targeting string

const (
	TargetingSelf   Targeting = "SELF"
	TargetingEnemy  Targeting = "ENEMY"
	TargetingBoth   Targeting = "BOTH"
)

// Ability represents a gopher's ability
type Ability struct {
	ID          string
	Name        string
	Description string
	Power       int
	Cost        int
	Targeting   Targeting
	EffectFunc  func(*BattleState, *Gopher, *Gopher) ([]string, error) // Returns log messages
}

// AbilityTemplate defines a template for creating abilities
type AbilityTemplate struct {
	Name        string
	Description string
	Power       int
	Cost        int
	Targeting   Targeting
}

// Predefined ability templates
var AbilityTemplates = map[string]AbilityTemplate{
	"quick_hit": {
		Name:        "Quick Hit",
		Description: "A fast, low-damage attack",
		Power:       20,
		Cost:        5,
		Targeting:   TargetingEnemy,
	},
	"go_panic": {
		Name:        "Go Panic()",
		Description: "Medium damage with chance to confuse",
		Power:       40,
		Cost:        10,
		Targeting:   TargetingEnemy,
	},
	"garbage_collector": {
		Name:        "Garbage Collector",
		Description: "Heal HP and cleanse status effects",
		Power:       30,
		Cost:        15,
		Targeting:   TargetingSelf,
	},
	"race_condition": {
		Name:        "Race Condition",
		Description: "High damage but can backfire",
		Power:       60,
		Cost:        20,
		Targeting:   TargetingEnemy,
	},
	"goroutine": {
		Name:        "Goroutine",
		Description: "Quick multi-hit attack",
		Power:       15,
		Cost:        8,
		Targeting:   TargetingEnemy,
	},
	"channel_blast": {
		Name:        "Channel Blast",
		Description: "Powerful channel-based attack",
		Power:       50,
		Cost:        18,
		Targeting:   TargetingEnemy,
	},
	"interface_guard": {
		Name:        "Interface Guard",
		Description: "Boost defense",
		Power:       25,
		Cost:        12,
		Targeting:   TargetingSelf,
	},
	"defer_recover": {
		Name:        "Defer Recover",
		Description: "Heal and prevent next attack",
		Power:       35,
		Cost:        20,
		Targeting:   TargetingSelf,
	},
}

// CreateAbilityFromTemplate creates an ability from a template
func CreateAbilityFromTemplate(templateID string, abilityID string) (*Ability, error) {
	template, ok := AbilityTemplates[templateID]
	if !ok {
		return nil, fmt.Errorf("unknown ability template: %s", templateID)
	}

	ability := &Ability{
		ID:          abilityID,
		Name:        template.Name,
		Description: template.Description,
		Power:       template.Power,
		Cost:        template.Cost,
		Targeting:   template.Targeting,
	}

	// Assign effect function based on template
	switch templateID {
	case "quick_hit":
		ability.EffectFunc = quickHitEffect
	case "go_panic":
		ability.EffectFunc = goPanicEffect
	case "garbage_collector":
		ability.EffectFunc = garbageCollectorEffect
	case "race_condition":
		ability.EffectFunc = raceConditionEffect
	case "goroutine":
		ability.EffectFunc = goroutineEffect
	case "channel_blast":
		ability.EffectFunc = channelBlastEffect
	case "interface_guard":
		ability.EffectFunc = interfaceGuardEffect
	case "defer_recover":
		ability.EffectFunc = deferRecoverEffect
	default:
		ability.EffectFunc = quickHitEffect // Default
	}

	return ability, nil
}

// Effect functions

func quickHitEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 20)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	return []string{fmt.Sprintf("%s used Quick Hit! Dealt %d damage!", user.Name, damage)}, nil
}

func goPanicEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 40)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Go Panic()! Dealt %d damage!", user.Name, damage)}
	
	// 30% chance to confuse
	if rand.Float64() < 0.3 {
		// Add confusion status (simplified for now)
		messages = append(messages, fmt.Sprintf("%s is confused!", target.Name))
	}
	
	return messages, nil
}

func garbageCollectorEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	heal := 30 + (user.Level * 2)
	user.CurrentHP += heal
	if user.CurrentHP > user.MaxHP {
		user.CurrentHP = user.MaxHP
	}
	return []string{fmt.Sprintf("%s used Garbage Collector! Healed %d HP!", user.Name, heal)}, nil
}

func raceConditionEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 60)
	
	messages := []string{}
	
	// 20% chance to backfire
	if rand.Float64() < 0.2 {
		user.CurrentHP -= damage / 2
		if user.CurrentHP < 0 {
			user.CurrentHP = 0
		}
		messages = append(messages, fmt.Sprintf("%s used Race Condition, but it backfired! Took %d damage!", user.Name, damage/2))
	} else {
		target.CurrentHP -= damage
		if target.CurrentHP < 0 {
			target.CurrentHP = 0
		}
		messages = append(messages, fmt.Sprintf("%s used Race Condition! Dealt %d damage!", user.Name, damage))
	}
	
	return messages, nil
}

func goroutineEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	hits := 2 + rand.Intn(2) // 2-3 hits
	totalDamage := 0
	
	for i := 0; i < hits; i++ {
		damage := calculateDamage(user, target, 15)
		target.CurrentHP -= damage
		totalDamage += damage
	}
	
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	return []string{fmt.Sprintf("%s used Goroutine! Hit %d times for %d total damage!", user.Name, hits, totalDamage)}, nil
}

func channelBlastEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 50)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	return []string{fmt.Sprintf("%s used Channel Blast! Dealt %d damage!", user.Name, damage)}, nil
}

func interfaceGuardEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	boost := 5 + (user.Level / 2)
	user.Defense += boost
	return []string{fmt.Sprintf("%s used Interface Guard! Defense increased by %d!", user.Name, boost)}, nil
}

func deferRecoverEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	heal := 35 + (user.Level * 3)
	user.CurrentHP += heal
	if user.CurrentHP > user.MaxHP {
		user.CurrentHP = user.MaxHP
	}
	// TODO: Add "prevent next attack" status
	return []string{fmt.Sprintf("%s used Defer Recover! Healed %d HP!", user.Name, heal)}, nil
}

// calculateDamage computes damage based on attacker and defender stats
func calculateDamage(attacker, defender *Gopher, basePower int) int {
	attack := float64(attacker.Attack) * (float64(basePower) / 100.0)
	defense := float64(defender.Defense) * 0.5
	
	damage := int(attack - defense)
	if damage < 1 {
		damage = 1
	}
	
	// Add some randomness (±10%)
	variance := float64(damage) * 0.1
	damage = int(float64(damage) + (rand.Float64()*2-1)*variance)
	
	if damage < 1 {
		damage = 1
	}
	
	return damage
}

