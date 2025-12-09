package game

import (
	"fmt"
	"math"
)

// PvPBattleState represents a PvP battle between two trainers
type PvPBattleState struct {
	ID                string
	ChannelID         string
	MessageID         string
	Trainer1ID        string
	Trainer2ID        string
	Trainer1Gopher    *Gopher
	Trainer2Gopher    *Gopher
	Trainer1Party     []*Gopher
	Trainer2Party     []*Gopher
	TurnOwner         string // "TRAINER1" or "TRAINER2"
	State             string // "PENDING", "ACTIVE", "TRAINER1_WON", "TRAINER2_WON", "DRAW"
	Log               []string
	EventManager      *EventManager
}

// NewPvPBattleState creates a new PvP battle state
func NewPvPBattleState(channelID string, trainer1ID, trainer2ID string, 
	trainer1Gopher, trainer2Gopher *Gopher, trainer1Party, trainer2Party []*Gopher,
	eventManager *EventManager) *PvPBattleState {
	
	return &PvPBattleState{
		ChannelID:      channelID,
		Trainer1ID:     trainer1ID,
		Trainer2ID:     trainer2ID,
		Trainer1Gopher: trainer1Gopher,
		Trainer2Gopher: trainer2Gopher,
		Trainer1Party:  trainer1Party,
		Trainer2Party:  trainer2Party,
		TurnOwner:      "TRAINER1", // Trainer who initiated goes first
		State:          "ACTIVE",
		Log:            []string{fmt.Sprintf("Battle between trainers started!")},
		EventManager:   eventManager,
	}
}

// CalculateELO calculates new ELO ratings after a battle
func CalculateELO(rating1, rating2 int, player1Won bool) (newRating1, newRating2 int) {
	k := 32 // K-factor
	
	expected1 := 1.0 / (1.0 + math.Pow(10.0, float64(rating2-rating1)/400.0))
	expected2 := 1.0 - expected1
	
	var score1, score2 float64
	if player1Won {
		score1 = 1.0
		score2 = 0.0
	} else {
		score1 = 0.0
		score2 = 1.0
	}
	
	newRating1 = rating1 + int(float64(k)*(score1-expected1))
	newRating2 = rating2 + int(float64(k)*(score2-expected2))
	
	return newRating1, newRating2
}

