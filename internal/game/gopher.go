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
	SpritePath      string
	GopherkonLayers []string
	IsInParty       bool
	PCSlot          *int
	Abilities       []*Ability
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
func GetAbilitiesForArchetype(archetype Archetype) []string {
	switch archetype {
	case ArchetypeHacker:
		return []string{"quick_hit", "go_panic", "goroutine", "race_condition"}
	case ArchetypeTank:
		return []string{"quick_hit", "interface_guard", "defer_recover", "garbage_collector"}
	case ArchetypeSpeedy:
		return []string{"quick_hit", "goroutine", "channel_blast", "go_panic"}
	case ArchetypeSupport:
		return []string{"garbage_collector", "interface_guard", "defer_recover", "quick_hit"}
	case ArchetypeMage:
		return []string{"channel_blast", "go_panic", "race_condition", "goroutine"}
	default:
		return []string{"quick_hit", "go_panic"}
	}
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

