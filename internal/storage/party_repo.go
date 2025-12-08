package storage

import (
	"fmt"
)

type PartyRepo struct {
	db          *DB
	gopherRepo  *GopherRepo
	trainerRepo *TrainerRepo
}

func NewPartyRepo(db *DB, gopherRepo *GopherRepo, trainerRepo *TrainerRepo) *PartyRepo {
	return &PartyRepo{
		db:          db,
		gopherRepo:  gopherRepo,
		trainerRepo: trainerRepo,
	}
}

func (r *PartyRepo) AddToParty(trainerID, gopherID string) error {
	party, err := r.gopherRepo.GetParty(trainerID)
	if err != nil {
		return fmt.Errorf("failed to get party: %w", err)
	}

	if len(party) >= 6 {
		return fmt.Errorf("party is full (max 6 gophers)")
	}

	gopher, err := r.gopherRepo.GetByID(gopherID)
	if err != nil {
		return fmt.Errorf("failed to get gopher: %w", err)
	}
	if gopher == nil {
		return fmt.Errorf("gopher not found")
	}

	if gopher.TrainerID == nil || *gopher.TrainerID != trainerID {
		return fmt.Errorf("gopher does not belong to trainer")
	}

	gopher.IsInParty = true
	gopher.PCSlot = nil

	if err := r.gopherRepo.Update(gopher); err != nil {
		return fmt.Errorf("failed to update gopher: %w", err)
	}

	// Update party slot count
	if err := r.trainerRepo.UpdatePartySlots(trainerID, len(party)+1); err != nil {
		return err
	}

	return nil
}

func (r *PartyRepo) RemoveFromParty(trainerID, gopherID string) error {
	party, err := r.gopherRepo.GetParty(trainerID)
	if err != nil {
		return fmt.Errorf("failed to get party: %w", err)
	}

	if len(party) <= 1 {
		return fmt.Errorf("cannot remove last gopher from party")
	}

	gopher, err := r.gopherRepo.GetByID(gopherID)
	if err != nil {
		return fmt.Errorf("failed to get gopher: %w", err)
	}
	if gopher == nil {
		return fmt.Errorf("gopher not found")
	}

	if gopher.TrainerID == nil || *gopher.TrainerID != trainerID {
		return fmt.Errorf("gopher does not belong to trainer")
	}

	// Find next available PC slot
	pcGophers, err := r.gopherRepo.GetPC(trainerID, 1000, 0)
	if err != nil {
		return fmt.Errorf("failed to get PC gophers: %w", err)
	}

	nextSlot := 0
	for _, pcGopher := range pcGophers {
		if pcGopher.PCSlot != nil && *pcGopher.PCSlot >= nextSlot {
			nextSlot = *pcGopher.PCSlot + 1
		}
	}

	gopher.IsInParty = false
	slot := nextSlot
	gopher.PCSlot = &slot

	if err := r.gopherRepo.Update(gopher); err != nil {
		return fmt.Errorf("failed to update gopher: %w", err)
	}

	// Update party slot count
	if err := r.trainerRepo.UpdatePartySlots(trainerID, len(party)-1); err != nil {
		return err
	}

	return nil
}

func (r *PartyRepo) GetPartySize(trainerID string) (int, error) {
	party, err := r.gopherRepo.GetParty(trainerID)
	if err != nil {
		return 0, err
	}
	return len(party), nil
}

