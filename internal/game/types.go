package game

import (
	"math/rand"
)

// GopherType represents a gopher's elemental type
type GopherType string

const (
	TypeHacker  GopherType = "Hacker"  // Tech/Code type
	TypeTank    GopherType = "Tank"    // Defense/Physical type
	TypeSpeedy  GopherType = "Speedy"  // Speed/Air type
	TypeSupport GopherType = "Support" // Healing/Support type
	TypeMage    GopherType = "Mage"    // Magic/Energy type
)

// AllTypes returns all available types
func AllTypes() []GopherType {
	return []GopherType{TypeHacker, TypeTank, TypeSpeedy, TypeSupport, TypeMage}
}

// GetTypeEffectiveness returns the effectiveness multiplier (0.5 = not very effective, 1.0 = normal, 2.0 = super effective)
func GetTypeEffectiveness(attackerType, defenderType GopherType) float64 {
	// Type effectiveness chart (like Pokemon)
	// Format: attacker > defender = effectiveness
	
	effectiveness := map[GopherType]map[GopherType]float64{
		TypeHacker: {
			TypeHacker:  1.0,  // Hacker vs Hacker: normal
			TypeTank:    0.5,  // Hacker vs Tank: not very effective (code can't break through armor)
			TypeSpeedy:  2.0,  // Hacker vs Speedy: super effective (code exploits speed)
			TypeSupport: 1.5,  // Hacker vs Support: effective (code disrupts support systems)
			TypeMage:    1.0,  // Hacker vs Mage: normal
		},
		TypeTank: {
			TypeHacker:  2.0,  // Tank vs Hacker: super effective (armor blocks code)
			TypeTank:    0.5,  // Tank vs Tank: not very effective (armor vs armor)
			TypeSpeedy:  0.5,  // Tank vs Speedy: not very effective (too slow to hit)
			TypeSupport: 1.5,  // Tank vs Support: effective (physical pressure)
			TypeMage:    1.5,  // Tank vs Mage: effective (armor resists magic)
		},
		TypeSpeedy: {
			TypeHacker:  1.5,  // Speedy vs Hacker: effective (speed outruns code)
			TypeTank:    2.0,  // Speedy vs Tank: super effective (speed bypasses armor)
			TypeSpeedy:  1.0,  // Speedy vs Speedy: normal
			TypeSupport: 1.0,  // Speedy vs Support: normal
			TypeMage:    1.5,  // Speedy vs Mage: effective (speed disrupts casting)
		},
		TypeSupport: {
			TypeHacker:  1.0,  // Support vs Hacker: normal
			TypeTank:    1.5,  // Support vs Tank: effective (support counters defense)
			TypeSpeedy:  1.0,  // Support vs Speedy: normal
			TypeSupport: 0.5,  // Support vs Support: not very effective (support cancels support)
			TypeMage:    2.0,  // Support vs Mage: super effective (support amplifies against magic)
		},
		TypeMage: {
			TypeHacker:  1.5,  // Mage vs Hacker: effective (magic disrupts code)
			TypeTank:    1.0,  // Mage vs Tank: normal (magic vs armor)
			TypeSpeedy:  1.5,  // Mage vs Speedy: effective (magic catches speed)
			TypeSupport: 0.5,  // Mage vs Support: not very effective (support counters magic)
			TypeMage:    1.0,  // Mage vs Mage: normal
		},
	}

	if chart, ok := effectiveness[attackerType]; ok {
		if mult, ok := chart[defenderType]; ok {
			return mult
		}
	}

	return 1.0 // Default: normal effectiveness
}

// GetDualTypeEffectiveness calculates effectiveness for dual-type gophers
func GetDualTypeEffectiveness(attackerType1, attackerType2 GopherType, defenderType1, defenderType2 GopherType) float64 {
	// For dual types, take the average of both matchups
	// If defender has no secondary type, use primary only
	if defenderType2 == "" {
		eff1 := GetTypeEffectiveness(attackerType1, defenderType1)
		eff2 := GetTypeEffectiveness(attackerType2, defenderType1)
		return (eff1 + eff2) / 2.0
	}

	// Both have dual types - average all combinations
	eff1 := GetTypeEffectiveness(attackerType1, defenderType1)
	eff2 := GetTypeEffectiveness(attackerType1, defenderType2)
	eff3 := GetTypeEffectiveness(attackerType2, defenderType1)
	eff4 := GetTypeEffectiveness(attackerType2, defenderType2)
	
	// Take the best matchup (like Pokemon)
	maxEff := eff1
	if eff2 > maxEff {
		maxEff = eff2
	}
	if eff3 > maxEff {
		maxEff = eff3
	}
	if eff4 > maxEff {
		maxEff = eff4
	}
	
	return maxEff
}

// GetTypeFromArchetype maps archetype to a primary type
func GetTypeFromArchetype(archetype Archetype) GopherType {
	switch archetype {
	case ArchetypeHacker:
		return TypeHacker
	case ArchetypeTank:
		return TypeTank
	case ArchetypeSpeedy:
		return TypeSpeedy
	case ArchetypeSupport:
		return TypeSupport
	case ArchetypeMage:
		return TypeMage
	default:
		return TypeHacker // Default
	}
}

// GetRandomSecondaryType returns a random secondary type (for dual-type gophers)
func GetRandomSecondaryType(primaryType GopherType) GopherType {
	allTypes := AllTypes()
	// Remove primary type from options
	options := []GopherType{}
	for _, t := range allTypes {
		if t != primaryType {
			options = append(options, t)
		}
	}
	
	if len(options) == 0 {
		return primaryType
	}
	
	return options[rand.Intn(len(options))]
}

// GetTypeEffectivenessMessage returns a message describing type effectiveness
func GetTypeEffectivenessMessage(effectiveness float64) string {
	if effectiveness >= 2.0 {
		return "It's super effective!"
	} else if effectiveness >= 1.5 {
		return "It's effective!"
	} else if effectiveness <= 0.5 {
		return "It's not very effective..."
	}
	return ""
}

