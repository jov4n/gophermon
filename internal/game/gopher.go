package game

import (
	"fmt"
	"math/rand"
)

// Archetype represents a gopher's class/type
type Archetype string

const (
	ArchetypeHacker  Archetype = "Hacker"
	ArchetypeTank    Archetype = "Tank"
	ArchetypeSpeedy  Archetype = "Speedy"
	ArchetypeSupport Archetype = "Support"
	ArchetypeMage    Archetype = "Mage"
)

// StatusEffectType represents different status conditions
type StatusEffectType string

const (
	StatusBurn       StatusEffectType = "BURN"
	StatusPoison     StatusEffectType = "POISON"
	StatusConfusion  StatusEffectType = "CONFUSION"
	StatusParalysis  StatusEffectType = "PARALYSIS"
	StatusSleep      StatusEffectType = "SLEEP"
	StatusAttackUp   StatusEffectType = "ATTACK_UP"
	StatusDefenseUp  StatusEffectType = "DEFENSE_UP"
	StatusSpeedUp    StatusEffectType = "SPEED_UP"
	StatusAttackDown StatusEffectType = "ATTACK_DOWN"
	StatusDefenseDown StatusEffectType = "DEFENSE_DOWN"
	StatusSpeedDown   StatusEffectType = "SPEED_DOWN"
	StatusProtect     StatusEffectType = "PROTECT" // Prevents next attack
)

// StatusEffect represents a status condition on a gopher
type StatusEffect struct {
	Type      StatusEffectType
	Duration  int // Turns remaining (-1 for permanent until removed)
	Intensity int // For damage over time effects
}

// Gopher represents a gopher with all its stats and abilities
type Gopher struct {
	ID              string
	TrainerID       *string
	Name            string
	Level           int
	XP              int
	CurrentHP       int
	MaxHP           int
	Attack          int
	Defense         int
	Speed           int
	Rarity          string
	ComplexityScore int
	SpeciesArchetype string
	EvolutionStage  int
	PrimaryType     GopherType  // Primary type (Hacker, Tank, Speedy, Support, Mage)
	SecondaryType   GopherType  // Secondary type for dual-type gophers (empty if single type)
	SpritePath      string  // Deprecated: kept for backward compatibility
	SpriteData      string  // Base64 encoded PNG image data
	GopherkonLayers []string
	IsInParty       bool
	PCSlot          *int
	Abilities       []*Ability
	StatusEffects   []*StatusEffect // Active status effects
	Shiny           bool            // Whether this gopher is shiny (rare color variant)
	IsFavorite      bool            // Whether this gopher is marked as favorite
	BaseAttack      int // Original attack before stat modifiers
	BaseDefense     int // Original defense before stat modifiers
	BaseSpeed       int // Original speed before stat modifiers
}

// XPNeeded calculates the XP required to reach a level
func XPNeeded(level int) int {
	return 50 * level * level
}

// AddXP adds experience and returns whether the gopher leveled up
func (g *Gopher) AddXP(amount int) (leveledUp bool, newLevel int) {
	g.XP += amount
	oldLevel := g.Level
	
	for g.XP >= XPNeeded(g.Level+1) {
		g.Level++
	}
	
	leveledUp = g.Level > oldLevel
	if leveledUp {
		g.LevelUp()
	}
	
	return leveledUp, g.Level
}

// LevelUp increases stats when leveling up
func (g *Gopher) LevelUp() {
	// HP increase
	hpIncrease := 10 + rand.Intn(6) + (g.ComplexityScore / 2)
	g.MaxHP += hpIncrease
	g.CurrentHP += hpIncrease // Heal on level up
	
	// Stat increases based on archetype
	switch Archetype(g.SpeciesArchetype) {
	case ArchetypeHacker:
		g.Attack += 3 + rand.Intn(3)
		g.Speed += 4 + rand.Intn(3)
		g.Defense += 1 + rand.Intn(2)
	case ArchetypeTank:
		g.MaxHP += 5 + rand.Intn(5) // Extra HP for tanks
		g.CurrentHP += 5 + rand.Intn(5)
		g.Defense += 4 + rand.Intn(3)
		g.Attack += 1 + rand.Intn(2)
		g.Speed += 1
	case ArchetypeSpeedy:
		g.Speed += 5 + rand.Intn(4)
		g.Attack += 2 + rand.Intn(3)
		g.Defense += 1 + rand.Intn(2)
	case ArchetypeSupport:
		g.Attack += 2 + rand.Intn(2)
		g.Defense += 2 + rand.Intn(2)
		g.Speed += 2 + rand.Intn(2)
		g.MaxHP += 3 + rand.Intn(3)
		g.CurrentHP += 3 + rand.Intn(3)
	case ArchetypeMage:
		g.Attack += 4 + rand.Intn(3)
		g.Defense += 2 + rand.Intn(2)
		g.Speed += 2 + rand.Intn(2)
	}
}

// GenerateBaseStats creates initial stats based on archetype and rarity
func GenerateBaseStats(archetype Archetype, rarity string, level int) (hp, attack, defense, speed int) {
	// Base stats by archetype
	var baseHP, baseAttack, baseDefense, baseSpeed int
	
	switch archetype {
	case ArchetypeHacker:
		baseHP = 60
		baseAttack = 45
		baseDefense = 30
		baseSpeed = 55
	case ArchetypeTank:
		baseHP = 90
		baseAttack = 35
		baseDefense = 50
		baseSpeed = 25
	case ArchetypeSpeedy:
		baseHP = 50
		baseAttack = 40
		baseDefense = 25
		baseSpeed = 65
	case ArchetypeSupport:
		baseHP = 70
		baseAttack = 35
		baseDefense = 40
		baseSpeed = 40
	case ArchetypeMage:
		baseHP = 55
		baseAttack = 50
		baseDefense = 30
		baseSpeed = 45
	default:
		baseHP = 60
		baseAttack = 40
		baseDefense = 35
		baseSpeed = 40
	}
	
	// Rarity multipliers
	rarityMultiplier := 1.0
	switch rarity {
	case "COMMON":
		rarityMultiplier = 1.0
	case "UNCOMMON":
		rarityMultiplier = 1.15
	case "RARE":
		rarityMultiplier = 1.3
	case "EPIC":
		rarityMultiplier = 1.5
	case "LEGENDARY":
		rarityMultiplier = 1.8
	}
	
	// Level scaling
	levelMultiplier := 1.0 + (float64(level-1) * 0.1)
	
	hp = int(float64(baseHP) * rarityMultiplier * levelMultiplier)
	attack = int(float64(baseAttack) * rarityMultiplier * levelMultiplier)
	defense = int(float64(baseDefense) * rarityMultiplier * levelMultiplier)
	speed = int(float64(baseSpeed) * rarityMultiplier * levelMultiplier)
	
	// Add some randomness (±5%)
	hp = int(float64(hp) * (0.95 + rand.Float64()*0.1))
	attack = int(float64(attack) * (0.95 + rand.Float64()*0.1))
	defense = int(float64(defense) * (0.95 + rand.Float64()*0.1))
	speed = int(float64(speed) * (0.95 + rand.Float64()*0.1))
	
	return hp, attack, defense, speed
}

// GetAbilitiesForArchetype returns ability template IDs for an archetype
// Includes base abilities, evolution abilities, and legendary abilities
func GetAbilitiesForArchetype(archetype Archetype, evolutionStage int, rarity string) []string {
	var baseAbilities []string
	
	switch archetype {
	case ArchetypeHacker:
		baseAbilities = []string{"quick_hit", "go_panic", "goroutine", "race_condition", "hack_attack", "burn_attack", "confuse_ray"}
	case ArchetypeTank:
		baseAbilities = []string{"quick_hit", "interface_guard", "defer_recover", "garbage_collector", "tank_slam", "harden", "break_armor"}
	case ArchetypeSpeedy:
		baseAbilities = []string{"quick_hit", "goroutine", "channel_blast", "go_panic", "speed_rush", "agility", "slow_down"}
	case ArchetypeSupport:
		baseAbilities = []string{"garbage_collector", "interface_guard", "defer_recover", "quick_hit", "support_boost", "poison_sting", "weaken"}
	case ArchetypeMage:
		baseAbilities = []string{"channel_blast", "go_panic", "race_condition", "goroutine", "magic_blast", "paralyze_bolt", "sleep_powder"}
	default:
		baseAbilities = []string{"quick_hit", "go_panic"}
	}
	
	// Add evolution stage 1 abilities
	if evolutionStage >= 1 {
		switch archetype {
		case ArchetypeHacker:
			baseAbilities = append(baseAbilities, "concurrent_strike", "mutex_lock")
		case ArchetypeTank:
			baseAbilities = append(baseAbilities, "reflect_guard", "context_timeout")
		case ArchetypeSpeedy:
			baseAbilities = append(baseAbilities, "select_storm", "agility")
		case ArchetypeSupport:
			baseAbilities = append(baseAbilities, "full_recovery", "power_up")
		case ArchetypeMage:
			baseAbilities = append(baseAbilities, "context_timeout", "magic_blast")
		}
	}
	
	// Add evolution stage 2 abilities
	if evolutionStage >= 2 {
		switch archetype {
		case ArchetypeHacker:
			baseAbilities = append(baseAbilities, "deadlock", "goroutine_swarm")
		case ArchetypeTank:
			baseAbilities = append(baseAbilities, "ultimate_guard", "channel_overload")
		case ArchetypeSpeedy:
			baseAbilities = append(baseAbilities, "goroutine_swarm", "channel_overload")
		case ArchetypeSupport:
			baseAbilities = append(baseAbilities, "full_recovery", "ultimate_guard")
		case ArchetypeMage:
			baseAbilities = append(baseAbilities, "channel_overload", "deadlock")
		}
	}
	
	// Add legendary abilities for legendary gophers
	if rarity == "LEGENDARY" {
		// All legendaries get at least one legendary ability
		legendaryAbilities := []string{"legendary_strike", "divine_heal", "god_mode"}
		// Randomly add 1-2 legendary abilities
		numLegendary := 1 + rand.Intn(2)
		for i := 0; i < numLegendary && i < len(legendaryAbilities); i++ {
			baseAbilities = append(baseAbilities, legendaryAbilities[i])
		}
		// Very rare chance for ultimate legendary abilities
		if rand.Float64() < 0.3 {
			if rand.Float64() < 0.5 {
				baseAbilities = append(baseAbilities, "apocalypse")
			} else {
				baseAbilities = append(baseAbilities, "time_rewind")
			}
		}
	}
	
	return baseAbilities
}

// GenerateGopherName creates a random name for a gopher
func GenerateGopherName(archetype Archetype) string {
	prefixes := []string{"Go", "Gopher", "Code", "Byte", "Bit", "Dev", "Hack"}
	suffixes := []string{"mon", "gopher", "coder", "dev", "hack", "byte", "bit"}
	
	if rand.Float64() < 0.3 {
		// Use archetype-based name
		return fmt.Sprintf("%s%s", archetype, suffixes[rand.Intn(len(suffixes))])
	}
	
	return fmt.Sprintf("%s%s", prefixes[rand.Intn(len(prefixes))], suffixes[rand.Intn(len(suffixes))])
}

// AddStatusEffect adds a status effect to the gopher
func (g *Gopher) AddStatusEffect(effectType StatusEffectType, duration int, intensity int) {
	// Check if effect already exists
	for _, effect := range g.StatusEffects {
		if effect.Type == effectType {
			// Refresh duration and update intensity if higher
			effect.Duration = duration
			if intensity > effect.Intensity {
				effect.Intensity = intensity
			}
			return
		}
	}
	
	// Add new effect
	g.StatusEffects = append(g.StatusEffects, &StatusEffect{
		Type:      effectType,
		Duration:  duration,
		Intensity: intensity,
	})
	
	// Store base stats on first status effect if not already stored
	if g.BaseAttack == 0 {
		g.BaseAttack = g.Attack
		g.BaseDefense = g.Defense
		g.BaseSpeed = g.Speed
	}
}

// RemoveStatusEffect removes a status effect
func (g *Gopher) RemoveStatusEffect(effectType StatusEffectType) {
	newEffects := []*StatusEffect{}
	for _, effect := range g.StatusEffects {
		if effect.Type != effectType {
			newEffects = append(newEffects, effect)
		}
	}
	g.StatusEffects = newEffects
	g.RecalculateStats()
}

// HasStatusEffect checks if gopher has a specific status effect
func (g *Gopher) HasStatusEffect(effectType StatusEffectType) bool {
	for _, effect := range g.StatusEffects {
		if effect.Type == effectType {
			return true
		}
	}
	return false
}

// RecalculateStats recalculates stats based on status effects
func (g *Gopher) RecalculateStats() {
	// Initialize base stats if not set
	if g.BaseAttack == 0 {
		g.BaseAttack = g.Attack
	}
	if g.BaseDefense == 0 {
		g.BaseDefense = g.Defense
	}
	if g.BaseSpeed == 0 {
		g.BaseSpeed = g.Speed
	}
	
	// Reset to base
	g.Attack = g.BaseAttack
	g.Defense = g.BaseDefense
	g.Speed = g.BaseSpeed
	
	// Apply stat modifiers
	for _, effect := range g.StatusEffects {
		switch effect.Type {
		case StatusAttackUp:
			g.Attack = int(float64(g.Attack) * 1.5)
		case StatusDefenseUp:
			g.Defense = int(float64(g.Defense) * 1.5)
		case StatusSpeedUp:
			g.Speed = int(float64(g.Speed) * 1.5)
		case StatusAttackDown:
			g.Attack = int(float64(g.Attack) * 0.75)
		case StatusDefenseDown:
			g.Defense = int(float64(g.Defense) * 0.75)
		case StatusSpeedDown:
			g.Speed = int(float64(g.Speed) * 0.75)
		}
	}
}

// ProcessStatusEffects processes status effects at the start of turn
// Returns messages about status effects
func (g *Gopher) ProcessStatusEffects() []string {
	messages := []string{}
	newEffects := []*StatusEffect{}
	
	for _, effect := range g.StatusEffects {
		// Decrease duration
		if effect.Duration > 0 {
			effect.Duration--
		}
		
		// Process effect
		switch effect.Type {
		case StatusBurn:
			damage := g.MaxHP / 8 // 12.5% max HP damage
			g.CurrentHP -= damage
			if g.CurrentHP < 0 {
				g.CurrentHP = 0
			}
			messages = append(messages, fmt.Sprintf("%s is hurt by burn! Lost %d HP!", g.Name, damage))
			if effect.Duration > 0 {
				newEffects = append(newEffects, effect)
			} else {
				messages = append(messages, fmt.Sprintf("%s is no longer burned!", g.Name))
			}
			
		case StatusPoison:
			damage := g.MaxHP / 16 + effect.Intensity // 6.25% + intensity
			g.CurrentHP -= damage
			if g.CurrentHP < 0 {
				g.CurrentHP = 0
			}
			messages = append(messages, fmt.Sprintf("%s is hurt by poison! Lost %d HP!", g.Name, damage))
			if effect.Duration > 0 {
				newEffects = append(newEffects, effect)
			} else {
				messages = append(messages, fmt.Sprintf("%s is no longer poisoned!", g.Name))
			}
			
		case StatusConfusion:
			// Confusion is handled in battle turn logic
			if effect.Duration > 0 {
				newEffects = append(newEffects, effect)
			} else {
				messages = append(messages, fmt.Sprintf("%s snapped out of confusion!", g.Name))
			}
			
		case StatusParalysis:
			// Paralysis is handled in battle turn logic
			if effect.Duration > 0 {
				newEffects = append(newEffects, effect)
			} else {
				messages = append(messages, fmt.Sprintf("%s is no longer paralyzed!", g.Name))
			}
			
		case StatusSleep:
			// Sleep is handled in battle turn logic
			if effect.Duration > 0 {
				newEffects = append(newEffects, effect)
			} else {
				messages = append(messages, fmt.Sprintf("%s woke up!", g.Name))
			}
			
		case StatusProtect:
			// Protect only lasts one turn
			// Don't add to newEffects
			messages = append(messages, fmt.Sprintf("%s's protection faded!", g.Name))
			
		default:
			// Stat modifiers persist
			if effect.Duration > 0 {
				newEffects = append(newEffects, effect)
			} else if effect.Duration == 0 {
				// Effect expired
				switch effect.Type {
				case StatusAttackUp:
					messages = append(messages, fmt.Sprintf("%s's attack returned to normal!", g.Name))
				case StatusDefenseUp:
					messages = append(messages, fmt.Sprintf("%s's defense returned to normal!", g.Name))
				case StatusSpeedUp:
					messages = append(messages, fmt.Sprintf("%s's speed returned to normal!", g.Name))
				case StatusAttackDown:
					messages = append(messages, fmt.Sprintf("%s's attack returned to normal!", g.Name))
				case StatusDefenseDown:
					messages = append(messages, fmt.Sprintf("%s's defense returned to normal!", g.Name))
				case StatusSpeedDown:
					messages = append(messages, fmt.Sprintf("%s's speed returned to normal!", g.Name))
				}
			}
		}
	}
	
	g.StatusEffects = newEffects
	g.RecalculateStats()
	
	return messages
}

