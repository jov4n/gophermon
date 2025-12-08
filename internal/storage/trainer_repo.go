package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Trainer struct {
	ID              string
	DiscordID       string
	Name            string
	CreatedAt       time.Time
	ActivePartySlots int
}

type TrainerRepo struct {
	db *DB
}

func NewTrainerRepo(db *DB) *TrainerRepo {
	return &TrainerRepo{db: db}
}

func (r *TrainerRepo) Create(discordID, name string) (*Trainer, error) {
	id := uuid.New().String()
	
	query := `INSERT INTO trainers (id, discord_id, name, active_party_slots) 
	          VALUES (?, ?, ?, 0)`
	
	_, err := r.db.Conn().Exec(query, id, discordID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create trainer: %w", err)
	}

	return r.GetByDiscordID(discordID)
}

func (r *TrainerRepo) GetByDiscordID(discordID string) (*Trainer, error) {
	query := `SELECT id, discord_id, name, created_at, active_party_slots 
	          FROM trainers WHERE discord_id = ?`
	
	var trainer Trainer
	var createdAt string
	
	err := r.db.Conn().QueryRow(query, discordID).Scan(
		&trainer.ID,
		&trainer.DiscordID,
		&trainer.Name,
		&createdAt,
		&trainer.ActivePartySlots,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trainer: %w", err)
	}

	trainer.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &trainer, nil
}

func (r *TrainerRepo) GetByID(id string) (*Trainer, error) {
	query := `SELECT id, discord_id, name, created_at, active_party_slots 
	          FROM trainers WHERE id = ?`
	
	var trainer Trainer
	var createdAt string
	
	err := r.db.Conn().QueryRow(query, id).Scan(
		&trainer.ID,
		&trainer.DiscordID,
		&trainer.Name,
		&createdAt,
		&trainer.ActivePartySlots,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trainer: %w", err)
	}

	trainer.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &trainer, nil
}

func (r *TrainerRepo) UpdatePartySlots(trainerID string, count int) error {
	query := `UPDATE trainers SET active_party_slots = ? WHERE id = ?`
	_, err := r.db.Conn().Exec(query, count, trainerID)
	return err
}

