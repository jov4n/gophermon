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
	Currency        int
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
	query := `SELECT id, discord_id, name, created_at, active_party_slots, COALESCE(currency, 100) 
	          FROM trainers WHERE discord_id = ?`
	
	var trainer Trainer
	var createdAt string
	
	err := r.db.Conn().QueryRow(query, discordID).Scan(
		&trainer.ID,
		&trainer.DiscordID,
		&trainer.Name,
		&createdAt,
		&trainer.ActivePartySlots,
		&trainer.Currency,
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
	query := `SELECT id, discord_id, name, created_at, active_party_slots, COALESCE(currency, 100) 
	          FROM trainers WHERE id = ?`
	
	var trainer Trainer
	var createdAt string
	
	err := r.db.Conn().QueryRow(query, id).Scan(
		&trainer.ID,
		&trainer.DiscordID,
		&trainer.Name,
		&createdAt,
		&trainer.ActivePartySlots,
		&trainer.Currency,
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

func (r *TrainerRepo) AddCurrency(trainerID string, amount int) error {
	query := `UPDATE trainers SET currency = COALESCE(currency, 100) + ? WHERE id = ?`
	_, err := r.db.Conn().Exec(query, amount, trainerID)
	return err
}

func (r *TrainerRepo) RemoveCurrency(trainerID string, amount int) error {
	query := `UPDATE trainers SET currency = COALESCE(currency, 100) - ? WHERE id = ? AND COALESCE(currency, 100) >= ?`
	result, err := r.db.Conn().Exec(query, amount, trainerID, amount)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient currency")
	}
	return nil
}

func (r *TrainerRepo) GetCurrency(trainerID string) (int, error) {
	var currency int
	err := r.db.Conn().QueryRow(
		"SELECT COALESCE(currency, 100) FROM trainers WHERE id = ?",
		trainerID,
	).Scan(&currency)
	return currency, err
}

