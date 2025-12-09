package storage

import (
	"fmt"
	"time"
)

type GopherdexEntry struct {
	TrainerID        string
	GopherName       string
	Archetype        string
	Rarity           string
	FirstEncounteredAt time.Time
	TimesEncountered int
	TimesCaught      int
	Owned            bool
}

type GopherdexRepo struct {
	db *DB
}

func NewGopherdexRepo(db *DB) *GopherdexRepo {
	return &GopherdexRepo{db: db}
}

func (r *GopherdexRepo) RecordEncounter(trainerID, gopherName, archetype, rarity string) error {
	// Check if entry exists
	var count int
	err := r.db.Conn().QueryRow(
		"SELECT COUNT(*) FROM gopherdex WHERE trainer_id = ? AND gopher_name = ? AND archetype = ? AND rarity = ?",
		trainerID, gopherName, archetype, rarity,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Update existing
		_, err = r.db.Conn().Exec(
			"UPDATE gopherdex SET times_encountered = times_encountered + 1 WHERE trainer_id = ? AND gopher_name = ? AND archetype = ? AND rarity = ?",
			trainerID, gopherName, archetype, rarity,
		)
	} else {
		// Insert new
		_, err = r.db.Conn().Exec(
			"INSERT INTO gopherdex (trainer_id, gopher_name, archetype, rarity, first_encountered_at, times_encountered) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 1)",
			trainerID, gopherName, archetype, rarity,
		)
	}
	return err
}

func (r *GopherdexRepo) RecordCatch(trainerID, gopherName, archetype, rarity string) error {
	// Check if entry exists
	var count int
	err := r.db.Conn().QueryRow(
		"SELECT COUNT(*) FROM gopherdex WHERE trainer_id = ? AND gopher_name = ? AND archetype = ? AND rarity = ?",
		trainerID, gopherName, archetype, rarity,
	).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		// Update existing
		_, err = r.db.Conn().Exec(
			"UPDATE gopherdex SET times_caught = times_caught + 1, owned = TRUE WHERE trainer_id = ? AND gopher_name = ? AND archetype = ? AND rarity = ?",
			trainerID, gopherName, archetype, rarity,
		)
	} else {
		// Insert new
		_, err = r.db.Conn().Exec(
			"INSERT INTO gopherdex (trainer_id, gopher_name, archetype, rarity, first_encountered_at, times_encountered, times_caught, owned) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 1, 1, TRUE)",
			trainerID, gopherName, archetype, rarity,
		)
	}
	return err
}

func (r *GopherdexRepo) SetOwned(trainerID, gopherName, archetype, rarity string, owned bool) error {
	query := `UPDATE gopherdex SET owned = ? WHERE trainer_id = ? AND gopher_name = ? AND archetype = ? AND rarity = ?`
	result, err := r.db.Conn().Exec(query, owned, trainerID, gopherName, archetype, rarity)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Entry doesn't exist, create it
		query = `INSERT INTO gopherdex (trainer_id, gopher_name, archetype, rarity, owned, first_encountered_at, times_encountered, times_caught)
		         VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 0, 0)`
		_, err = r.db.Conn().Exec(query, trainerID, gopherName, archetype, rarity, owned)
	}
	return err
}

func (r *GopherdexRepo) GetEntries(trainerID string) ([]*GopherdexEntry, error) {
	rows, err := r.db.Conn().Query(
		`SELECT trainer_id, gopher_name, archetype, rarity, first_encountered_at, 
		 times_encountered, times_caught, owned
		 FROM gopherdex WHERE trainer_id = ?`,
		trainerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query gopherdex: %w", err)
	}
	defer rows.Close()

	var entries []*GopherdexEntry
	for rows.Next() {
		entry := &GopherdexEntry{}
		var firstEncountered string
		if err := rows.Scan(&entry.TrainerID, &entry.GopherName, &entry.Archetype, &entry.Rarity,
			&firstEncountered, &entry.TimesEncountered, &entry.TimesCaught, &entry.Owned); err != nil {
			return nil, fmt.Errorf("failed to scan gopherdex entry: %w", err)
		}
		entry.FirstEncounteredAt, _ = time.Parse("2006-01-02 15:04:05", firstEncountered)
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *GopherdexRepo) GetCompletion(trainerID string) (int, int, error) {
	var total, owned int
	err := r.db.Conn().QueryRow(
		`SELECT COUNT(*), SUM(CASE WHEN owned = TRUE THEN 1 ELSE 0 END) 
		 FROM gopherdex WHERE trainer_id = ?`,
		trainerID,
	).Scan(&total, &owned)
	return owned, total, err
}

