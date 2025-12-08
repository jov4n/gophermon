package game

// Rarity represents the rarity tier of a gopher
type Rarity string

const (
	RarityCommon    Rarity = "COMMON"
	RarityUncommon  Rarity = "UNCOMMON"
	RarityRare      Rarity = "RARE"
	RarityEpic      Rarity = "EPIC"
	RarityLegendary Rarity = "LEGENDARY"
)

// ComplexityToRarity maps complexity score to rarity tier
func ComplexityToRarity(complexity int) Rarity {
	switch {
	case complexity <= 2:
		return RarityCommon
	case complexity <= 4:
		return RarityUncommon
	case complexity <= 6:
		return RarityRare
	case complexity <= 8:
		return RarityEpic
	default:
		return RarityLegendary
	}
}

// RarityToComplexityRange returns the min and max complexity for a rarity tier
func RarityToComplexityRange(rarity Rarity) (min, max int) {
	switch rarity {
	case RarityCommon:
		return 1, 2
	case RarityUncommon:
		return 3, 4
	case RarityRare:
		return 5, 6
	case RarityEpic:
		return 7, 8
	case RarityLegendary:
		return 9, 15
	default:
		return 1, 2
	}
}

// String returns the string representation of Rarity
func (r Rarity) String() string {
	return string(r)
}

// GetWildRarityDistribution returns a rarity based on wild encounter distribution
// 60% Common, 25% Uncommon, 10% Rare, 4% Epic, 1% Legendary
func GetWildRarityDistribution(randFloat float64) Rarity {
	switch {
	case randFloat < 0.60:
		return RarityCommon
	case randFloat < 0.85:
		return RarityUncommon
	case randFloat < 0.95:
		return RarityRare
	case randFloat < 0.99:
		return RarityEpic
	default:
		return RarityLegendary
	}
}

