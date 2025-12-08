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
	// Status effect abilities
	"burn_attack": {
		Name:        "Flame On",
		Description: "Attack that may burn the target",
		Power:       35,
		Cost:        12,
		Targeting:   TargetingEnemy,
	},
	"poison_sting": {
		Name:        "Toxic Code",
		Description: "Attack that may poison the target",
		Power:       30,
		Cost:        10,
		Targeting:   TargetingEnemy,
	},
	"paralyze_bolt": {
		Name:        "Static Shock",
		Description: "Attack that may paralyze the target",
		Power:       35,
		Cost:        12,
		Targeting:   TargetingEnemy,
	},
	"sleep_powder": {
		Name:        "Sleep Mode",
		Description: "May put the target to sleep",
		Power:       0,
		Cost:        15,
		Targeting:   TargetingEnemy,
	},
	"confuse_ray": {
		Name:        "Confuse Ray",
		Description: "Confuses the target",
		Power:       0,
		Cost:        10,
		Targeting:   TargetingEnemy,
	},
	// Buff abilities
	"power_up": {
		Name:        "Power Up",
		Description: "Increases attack for several turns",
		Power:       0,
		Cost:        15,
		Targeting:   TargetingSelf,
	},
	"harden": {
		Name:        "Harden",
		Description: "Increases defense for several turns",
		Power:       0,
		Cost:        15,
		Targeting:   TargetingSelf,
	},
	"agility": {
		Name:        "Agility",
		Description: "Increases speed for several turns",
		Power:       0,
		Cost:        15,
		Targeting:   TargetingSelf,
	},
	// Debuff abilities
	"weaken": {
		Name:        "Weaken",
		Description: "Reduces enemy attack",
		Power:       0,
		Cost:        12,
		Targeting:   TargetingEnemy,
	},
	"break_armor": {
		Name:        "Break Armor",
		Description: "Reduces enemy defense",
		Power:       0,
		Cost:        12,
		Targeting:   TargetingEnemy,
	},
	"slow_down": {
		Name:        "Slow Down",
		Description: "Reduces enemy speed",
		Power:       0,
		Cost:        12,
		Targeting:   TargetingEnemy,
	},
	// Type-specific abilities
	"hack_attack": {
		Name:        "Hack Attack",
		Description: "Powerful Hacker-type attack",
		Power:       55,
		Cost:        18,
		Targeting:   TargetingEnemy,
	},
	"tank_slam": {
		Name:        "Tank Slam",
		Description: "Powerful Tank-type attack",
		Power:       55,
		Cost:        18,
		Targeting:   TargetingEnemy,
	},
	"speed_rush": {
		Name:        "Speed Rush",
		Description: "Powerful Speedy-type attack",
		Power:       55,
		Cost:        18,
		Targeting:   TargetingEnemy,
	},
	"support_boost": {
		Name:        "Support Boost",
		Description: "Heals and boosts stats",
		Power:       40,
		Cost:        20,
		Targeting:   TargetingSelf,
	},
	"magic_blast": {
		Name:        "Magic Blast",
		Description: "Powerful Mage-type attack",
		Power:       55,
		Cost:        18,
		Targeting:   TargetingEnemy,
	},
	// Evolution stage 1 abilities (unlocked at evolution stage 1)
	"concurrent_strike": {
		Name:        "Concurrent Strike",
		Description: "Multi-hit attack (Evolution Stage 1+)",
		Power:       25,
		Cost:        15,
		Targeting:   TargetingEnemy,
	},
	"mutex_lock": {
		Name:        "Mutex Lock",
		Description: "Powerful attack that may paralyze (Evolution Stage 1+)",
		Power:       65,
		Cost:        22,
		Targeting:   TargetingEnemy,
	},
	"context_timeout": {
		Name:        "Context Timeout",
		Description: "High damage with chance to sleep (Evolution Stage 1+)",
		Power:       60,
		Cost:        20,
		Targeting:   TargetingEnemy,
	},
	"reflect_guard": {
		Name:        "Reflect Guard",
		Description: "Boost defense and reflect damage (Evolution Stage 1+)",
		Power:       0,
		Cost:        18,
		Targeting:   TargetingSelf,
	},
	"select_storm": {
		Name:        "Select Storm",
		Description: "Multi-hit channel attack (Evolution Stage 1+)",
		Power:       20,
		Cost:        16,
		Targeting:   TargetingEnemy,
	},
	// Evolution stage 2 abilities (unlocked at evolution stage 2)
	"deadlock": {
		Name:        "Deadlock",
		Description: "Devastating attack that may freeze both (Evolution Stage 2+)",
		Power:       80,
		Cost:        30,
		Targeting:   TargetingEnemy,
	},
	"goroutine_swarm": {
		Name:        "Goroutine Swarm",
		Description: "Massive multi-hit attack (Evolution Stage 2+)",
		Power:       18,
		Cost:        25,
		Targeting:   TargetingEnemy,
	},
	"channel_overload": {
		Name:        "Channel Overload",
		Description: "Extremely powerful channel attack (Evolution Stage 2+)",
		Power:       90,
		Cost:        35,
		Targeting:   TargetingEnemy,
	},
	"full_recovery": {
		Name:        "Full Recovery",
		Description: "Full heal and status cleanse (Evolution Stage 2+)",
		Power:       0,
		Cost:        30,
		Targeting:   TargetingSelf,
	},
	"ultimate_guard": {
		Name:        "Ultimate Guard",
		Description: "Massive defense boost and protection (Evolution Stage 2+)",
		Power:       0,
		Cost:        28,
		Targeting:   TargetingSelf,
	},
	// Legendary-specific abilities
	"legendary_strike": {
		Name:        "Legendary Strike",
		Description: "Legendary gopher's signature attack",
		Power:       100,
		Cost:        40,
		Targeting:   TargetingEnemy,
	},
	"divine_heal": {
		Name:        "Divine Heal",
		Description: "Legendary healing that restores all HP",
		Power:       0,
		Cost:        35,
		Targeting:   TargetingSelf,
	},
	"god_mode": {
		Name:        "God Mode",
		Description: "Legendary buff that boosts all stats",
		Power:       0,
		Cost:        40,
		Targeting:   TargetingSelf,
	},
	"apocalypse": {
		Name:        "Apocalypse",
		Description: "Legendary attack that damages both fighters",
		Power:       120,
		Cost:        50,
		Targeting:   TargetingBoth,
	},
	"time_rewind": {
		Name:        "Time Rewind",
		Description: "Legendary ability that fully restores HP and removes all status",
		Power:       0,
		Cost:        45,
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
	// Status effect abilities
	case "burn_attack":
		ability.EffectFunc = burnAttackEffect
	case "poison_sting":
		ability.EffectFunc = poisonStingEffect
	case "paralyze_bolt":
		ability.EffectFunc = paralyzeBoltEffect
	case "sleep_powder":
		ability.EffectFunc = sleepPowderEffect
	case "confuse_ray":
		ability.EffectFunc = confuseRayEffect
	// Buff abilities
	case "power_up":
		ability.EffectFunc = powerUpEffect
	case "harden":
		ability.EffectFunc = hardenEffect
	case "agility":
		ability.EffectFunc = agilityEffect
	// Debuff abilities
	case "weaken":
		ability.EffectFunc = weakenEffect
	case "break_armor":
		ability.EffectFunc = breakArmorEffect
	case "slow_down":
		ability.EffectFunc = slowDownEffect
	// Type-specific abilities
	case "hack_attack":
		ability.EffectFunc = hackAttackEffect
	case "tank_slam":
		ability.EffectFunc = tankSlamEffect
	case "speed_rush":
		ability.EffectFunc = speedRushEffect
	case "support_boost":
		ability.EffectFunc = supportBoostEffect
	case "magic_blast":
		ability.EffectFunc = magicBlastEffect
	// Evolution stage 1 abilities
	case "concurrent_strike":
		ability.EffectFunc = concurrentStrikeEffect
	case "mutex_lock":
		ability.EffectFunc = mutexLockEffect
	case "context_timeout":
		ability.EffectFunc = contextTimeoutEffect
	case "reflect_guard":
		ability.EffectFunc = reflectGuardEffect
	case "select_storm":
		ability.EffectFunc = selectStormEffect
	// Evolution stage 2 abilities
	case "deadlock":
		ability.EffectFunc = deadlockEffect
	case "goroutine_swarm":
		ability.EffectFunc = goroutineSwarmEffect
	case "channel_overload":
		ability.EffectFunc = channelOverloadEffect
	case "full_recovery":
		ability.EffectFunc = fullRecoveryEffect
	case "ultimate_guard":
		ability.EffectFunc = ultimateGuardEffect
	// Legendary abilities
	case "legendary_strike":
		ability.EffectFunc = legendaryStrikeEffect
	case "divine_heal":
		ability.EffectFunc = divineHealEffect
	case "god_mode":
		ability.EffectFunc = godModeEffect
	case "apocalypse":
		ability.EffectFunc = apocalypseEffect
	case "time_rewind":
		ability.EffectFunc = timeRewindEffect
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
	
	messages := []string{fmt.Sprintf("%s used Quick Hit! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := 1.0
	if user.SecondaryType != "" {
		effectiveness = GetDualTypeEffectiveness(user.PrimaryType, user.SecondaryType, target.PrimaryType, target.SecondaryType)
	} else {
		if target.SecondaryType != "" {
			eff1 := GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
			eff2 := GetTypeEffectiveness(user.PrimaryType, target.SecondaryType)
			effectiveness = (eff1 + eff2) / 2.0
		} else {
			effectiveness = GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
		}
	}
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func goPanicEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 40)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Go Panic()! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := 1.0
	if user.SecondaryType != "" {
		effectiveness = GetDualTypeEffectiveness(user.PrimaryType, user.SecondaryType, target.PrimaryType, target.SecondaryType)
	} else {
		if target.SecondaryType != "" {
			eff1 := GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
			eff2 := GetTypeEffectiveness(user.PrimaryType, target.SecondaryType)
			effectiveness = (eff1 + eff2) / 2.0
		} else {
			effectiveness = GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
		}
	}
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	// 30% chance to confuse
	if rand.Float64() < 0.3 {
		duration := 2 + rand.Intn(2)
		target.AddStatusEffect(StatusConfusion, duration, 0)
		messages = append(messages, fmt.Sprintf("%s became confused!", target.Name))
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
	user.AddStatusEffect(StatusProtect, 1, 0)
	return []string{fmt.Sprintf("%s used Defer Recover! Healed %d HP and gained protection!", user.Name, heal)}, nil
}

// Status effect abilities
func burnAttackEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 35)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Flame On! Dealt %d damage!", user.Name, damage)}
	
	// 40% chance to burn
	if rand.Float64() < 0.4 {
		target.AddStatusEffect(StatusBurn, 3+rand.Intn(2), 0)
		messages = append(messages, fmt.Sprintf("%s was burned!", target.Name))
	}
	
	return messages, nil
}

func poisonStingEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 30)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Toxic Code! Dealt %d damage!", user.Name, damage)}
	
	// 50% chance to poison
	if rand.Float64() < 0.5 {
		intensity := 2 + (user.Level / 5)
		target.AddStatusEffect(StatusPoison, 4+rand.Intn(2), intensity)
		messages = append(messages, fmt.Sprintf("%s was poisoned!", target.Name))
	}
	
	return messages, nil
}

func paralyzeBoltEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 35)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Static Shock! Dealt %d damage!", user.Name, damage)}
	
	// 30% chance to paralyze
	if rand.Float64() < 0.3 {
		target.AddStatusEffect(StatusParalysis, 2+rand.Intn(2), 0)
		messages = append(messages, fmt.Sprintf("%s was paralyzed!", target.Name))
	}
	
	return messages, nil
}

func sleepPowderEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	messages := []string{fmt.Sprintf("%s used Sleep Mode!", user.Name)}
	
	// 60% chance to sleep
	if rand.Float64() < 0.6 {
		duration := 2 + rand.Intn(2)
		target.AddStatusEffect(StatusSleep, duration, 0)
		messages = append(messages, fmt.Sprintf("%s fell asleep!", target.Name))
	} else {
		messages = append(messages, fmt.Sprintf("But it failed!"))
	}
	
	return messages, nil
}

func confuseRayEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	messages := []string{fmt.Sprintf("%s used Confuse Ray!", user.Name)}
	
	// 70% chance to confuse
	if rand.Float64() < 0.7 {
		duration := 2 + rand.Intn(2)
		target.AddStatusEffect(StatusConfusion, duration, 0)
		messages = append(messages, fmt.Sprintf("%s became confused!", target.Name))
	} else {
		messages = append(messages, fmt.Sprintf("But it failed!"))
	}
	
	return messages, nil
}

// Buff abilities
func powerUpEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	user.AddStatusEffect(StatusAttackUp, duration, 0)
	user.RecalculateStats()
	return []string{fmt.Sprintf("%s used Power Up! Attack increased!", user.Name)}, nil
}

func hardenEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	user.AddStatusEffect(StatusDefenseUp, duration, 0)
	user.RecalculateStats()
	return []string{fmt.Sprintf("%s used Harden! Defense increased!", user.Name)}, nil
}

func agilityEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	user.AddStatusEffect(StatusSpeedUp, duration, 0)
	user.RecalculateStats()
	return []string{fmt.Sprintf("%s used Agility! Speed increased!", user.Name)}, nil
}

// Debuff abilities
func weakenEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	target.AddStatusEffect(StatusAttackDown, duration, 0)
	target.RecalculateStats()
	return []string{fmt.Sprintf("%s used Weaken! %s's attack decreased!", user.Name, target.Name)}, nil
}

func breakArmorEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	target.AddStatusEffect(StatusDefenseDown, duration, 0)
	target.RecalculateStats()
	return []string{fmt.Sprintf("%s used Break Armor! %s's defense decreased!", user.Name, target.Name)}, nil
}

func slowDownEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	target.AddStatusEffect(StatusSpeedDown, duration, 0)
	target.RecalculateStats()
	return []string{fmt.Sprintf("%s used Slow Down! %s's speed decreased!", user.Name, target.Name)}, nil
}

// Type-specific abilities
func hackAttackEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 55)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Hack Attack! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := GetTypeEffectiveness(TypeHacker, target.PrimaryType)
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func tankSlamEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 55)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Tank Slam! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := GetTypeEffectiveness(TypeTank, target.PrimaryType)
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func speedRushEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 55)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Speed Rush! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := GetTypeEffectiveness(TypeSpeedy, target.PrimaryType)
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func supportBoostEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	heal := 40 + (user.Level * 3)
	user.CurrentHP += heal
	if user.CurrentHP > user.MaxHP {
		user.CurrentHP = user.MaxHP
	}
	
	duration := 3 + (user.Level / 5)
	user.AddStatusEffect(StatusAttackUp, duration, 0)
	user.AddStatusEffect(StatusDefenseUp, duration, 0)
	user.RecalculateStats()
	
	return []string{fmt.Sprintf("%s used Support Boost! Healed %d HP and boosted stats!", user.Name, heal)}, nil
}

func magicBlastEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 55)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Magic Blast! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := GetTypeEffectiveness(TypeMage, target.PrimaryType)
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

// Evolution stage 1 abilities
func concurrentStrikeEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	hits := 3 + rand.Intn(2) // 3-4 hits
	totalDamage := 0
	
	for i := 0; i < hits; i++ {
		damage := calculateDamage(user, target, 25)
		target.CurrentHP -= damage
		totalDamage += damage
	}
	
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	return []string{fmt.Sprintf("%s used Concurrent Strike! Hit %d times for %d total damage!", user.Name, hits, totalDamage)}, nil
}

func mutexLockEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 65)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Mutex Lock! Dealt %d damage!", user.Name, damage)}
	
	// 40% chance to paralyze
	if rand.Float64() < 0.4 {
		target.AddStatusEffect(StatusParalysis, 2+rand.Intn(2), 0)
		messages = append(messages, fmt.Sprintf("%s was paralyzed!", target.Name))
	}
	
	return messages, nil
}

func contextTimeoutEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 60)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Context Timeout! Dealt %d damage!", user.Name, damage)}
	
	// 35% chance to sleep
	if rand.Float64() < 0.35 {
		duration := 2 + rand.Intn(2)
		target.AddStatusEffect(StatusSleep, duration, 0)
		messages = append(messages, fmt.Sprintf("%s fell asleep!", target.Name))
	}
	
	return messages, nil
}

func reflectGuardEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 3 + (user.Level / 5)
	user.AddStatusEffect(StatusDefenseUp, duration, 0)
	user.AddStatusEffect(StatusProtect, 1, 0)
	user.RecalculateStats()
	return []string{fmt.Sprintf("%s used Reflect Guard! Defense increased and protection gained!", user.Name)}, nil
}

func selectStormEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	hits := 4 + rand.Intn(2) // 4-5 hits
	totalDamage := 0
	
	for i := 0; i < hits; i++ {
		damage := calculateDamage(user, target, 20)
		target.CurrentHP -= damage
		totalDamage += damage
	}
	
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	return []string{fmt.Sprintf("%s used Select Storm! Hit %d times for %d total damage!", user.Name, hits, totalDamage)}, nil
}

// Evolution stage 2 abilities
func deadlockEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 80)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Deadlock! Dealt %d damage!", user.Name, damage)}
	
	// 30% chance to paralyze both
	if rand.Float64() < 0.3 {
		user.AddStatusEffect(StatusParalysis, 1, 0)
		target.AddStatusEffect(StatusParalysis, 1, 0)
		messages = append(messages, "Both fighters were paralyzed by the deadlock!")
	}
	
	return messages, nil
}

func goroutineSwarmEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	hits := 5 + rand.Intn(3) // 5-7 hits
	totalDamage := 0
	
	for i := 0; i < hits; i++ {
		damage := calculateDamage(user, target, 18)
		target.CurrentHP -= damage
		totalDamage += damage
	}
	
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	return []string{fmt.Sprintf("%s used Goroutine Swarm! Hit %d times for %d total damage!", user.Name, hits, totalDamage)}, nil
}

func channelOverloadEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 90)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Channel Overload! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := 1.0
	if user.SecondaryType != "" {
		effectiveness = GetDualTypeEffectiveness(user.PrimaryType, user.SecondaryType, target.PrimaryType, target.SecondaryType)
	} else {
		if target.SecondaryType != "" {
			eff1 := GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
			eff2 := GetTypeEffectiveness(user.PrimaryType, target.SecondaryType)
			effectiveness = (eff1 + eff2) / 2.0
		} else {
			effectiveness = GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
		}
	}
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func fullRecoveryEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	user.CurrentHP = user.MaxHP
	
	// Remove all negative status effects
	newEffects := []*StatusEffect{}
	for _, effect := range user.StatusEffects {
		if effect.Type == StatusAttackUp || effect.Type == StatusDefenseUp || effect.Type == StatusSpeedUp {
			newEffects = append(newEffects, effect)
		}
	}
	user.StatusEffects = newEffects
	
	return []string{fmt.Sprintf("%s used Full Recovery! Fully restored HP and cleansed negative status!", user.Name)}, nil
}

func ultimateGuardEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 4 + (user.Level / 5)
	user.AddStatusEffect(StatusDefenseUp, duration, 0)
	user.AddStatusEffect(StatusProtect, 1, 0)
	user.RecalculateStats()
	return []string{fmt.Sprintf("%s used Ultimate Guard! Massive defense boost and protection gained!", user.Name)}, nil
}

// Legendary abilities
func legendaryStrikeEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 100)
	target.CurrentHP -= damage
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Legendary Strike! Dealt %d damage!", user.Name, damage)}
	
	// Add type effectiveness message
	effectiveness := 1.0
	if user.SecondaryType != "" {
		effectiveness = GetDualTypeEffectiveness(user.PrimaryType, user.SecondaryType, target.PrimaryType, target.SecondaryType)
	} else {
		if target.SecondaryType != "" {
			eff1 := GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
			eff2 := GetTypeEffectiveness(user.PrimaryType, target.SecondaryType)
			effectiveness = (eff1 + eff2) / 2.0
		} else {
			effectiveness = GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
		}
	}
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func divineHealEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	user.CurrentHP = user.MaxHP
	return []string{fmt.Sprintf("%s used Divine Heal! Fully restored HP!", user.Name)}, nil
}

func godModeEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	duration := 5 + (user.Level / 3)
	user.AddStatusEffect(StatusAttackUp, duration, 0)
	user.AddStatusEffect(StatusDefenseUp, duration, 0)
	user.AddStatusEffect(StatusSpeedUp, duration, 0)
	user.RecalculateStats()
	return []string{fmt.Sprintf("%s activated God Mode! All stats massively increased!", user.Name)}, nil
}

func apocalypseEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	damage := calculateDamage(user, target, 120)
	
	// Damage both
	user.CurrentHP -= damage / 3
	target.CurrentHP -= damage
	if user.CurrentHP < 0 {
		user.CurrentHP = 0
	}
	if target.CurrentHP < 0 {
		target.CurrentHP = 0
	}
	
	messages := []string{fmt.Sprintf("%s used Apocalypse! Dealt %d damage to %s and %d to itself!", user.Name, damage, target.Name, damage/3)}
	
	// Add type effectiveness message
	effectiveness := 1.0
	if user.SecondaryType != "" {
		effectiveness = GetDualTypeEffectiveness(user.PrimaryType, user.SecondaryType, target.PrimaryType, target.SecondaryType)
	} else {
		if target.SecondaryType != "" {
			eff1 := GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
			eff2 := GetTypeEffectiveness(user.PrimaryType, target.SecondaryType)
			effectiveness = (eff1 + eff2) / 2.0
		} else {
			effectiveness = GetTypeEffectiveness(user.PrimaryType, target.PrimaryType)
		}
	}
	if msg := GetTypeEffectivenessMessage(effectiveness); msg != "" {
		messages = append(messages, msg)
	}
	
	return messages, nil
}

func timeRewindEffect(state *BattleState, user, target *Gopher) ([]string, error) {
	user.CurrentHP = user.MaxHP
	
	// Remove all status effects
	user.StatusEffects = []*StatusEffect{}
	user.RecalculateStats()
	
	return []string{fmt.Sprintf("%s used Time Rewind! Fully restored HP and removed all status effects!", user.Name)}, nil
}

// calculateDamage computes damage based on attacker and defender stats, including type effectiveness
func calculateDamage(attacker, defender *Gopher, basePower int) int {
	attack := float64(attacker.Attack) * (float64(basePower) / 100.0)
	defense := float64(defender.Defense) * 0.5
	
	damage := int(attack - defense)
	if damage < 1 {
		damage = 1
	}
	
	// Apply type effectiveness
	effectiveness := 1.0
	if attacker.SecondaryType != "" {
		// Dual type attacker
		effectiveness = GetDualTypeEffectiveness(
			attacker.PrimaryType, attacker.SecondaryType,
			defender.PrimaryType, defender.SecondaryType,
		)
	} else {
		// Single type attacker
		if defender.SecondaryType != "" {
			// Defender has dual type - average effectiveness
			eff1 := GetTypeEffectiveness(attacker.PrimaryType, defender.PrimaryType)
			eff2 := GetTypeEffectiveness(attacker.PrimaryType, defender.SecondaryType)
			effectiveness = (eff1 + eff2) / 2.0
		} else {
			// Both single type
			effectiveness = GetTypeEffectiveness(attacker.PrimaryType, defender.PrimaryType)
		}
	}
	
	damage = int(float64(damage) * effectiveness)
	
	// Add some randomness (±10%)
	variance := float64(damage) * 0.1
	damage = int(float64(damage) + (rand.Float64()*2-1)*variance)
	
	if damage < 1 {
		damage = 1
	}
	
	return damage
}

