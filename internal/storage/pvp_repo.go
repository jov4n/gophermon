package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type PvPStats struct {
	TrainerID    string
	Wins          int
	Losses        int
	Draws         int
	Rating        int
	HighestRating int
	TotalBattles  int
	UpdatedAt     time.Time
}

type PvPRepo struct {
	db *DB
}

func NewPvPRepo(db *DB) *PvPRepo {
	return &PvPRepo{db: db}
}

func (r *PvPRepo) GetOrCreate(trainerID string) (*PvPStats, error) {
	query := `SELECT trainer_id, wins, losses, draws, rating, highest_rating, total_battles, updated_at
	          FROM pvp_stats WHERE trainer_id = ?`
	
	var stats PvPStats
	var updatedAt string
	
	err := r.db.Conn().QueryRow(query, trainerID).Scan(
		&stats.TrainerID, &stats.Wins, &stats.Losses, &stats.Draws,
		&stats.Rating, &stats.HighestRating, &stats.TotalBattles, &updatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Create new stats
		_, err = r.db.Conn().Exec(
			`INSERT INTO pvp_stats (trainer_id, wins, losses, draws, rating, highest_rating, total_battles) 
			 VALUES (?, 0, 0, 0, 1000, 1000, 0)`,
			trainerID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create pvp stats: %w", err)
		}
		return r.GetOrCreate(trainerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pvp stats: %w", err)
	}

	stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &stats, nil
}

func (r *PvPRepo) UpdateRating(trainerID string, newRating int, won bool) error {
	stats, err := r.GetOrCreate(trainerID)
	if err != nil {
		return err
	}

	var wins, losses, draws int
	if won {
		wins = 1
	} else {
		losses = 1
	}

	highestRating := stats.HighestRating
	if newRating > highestRating {
		highestRating = newRating
	}

	_, err = r.db.Conn().Exec(
		`UPDATE pvp_stats SET wins = wins + ?, losses = losses + ?, draws = draws + ?,
		 rating = ?, highest_rating = ?, total_battles = total_battles + 1, updated_at = CURRENT_TIMESTAMP
		 WHERE trainer_id = ?`,
		wins, losses, draws, newRating, highestRating, trainerID,
	)
	return err
}

func (r *PvPRepo) GetLeaderboard(limit int) ([]*PvPStats, error) {
	rows, err := r.db.Conn().Query(
		`SELECT trainer_id, wins, losses, draws, rating, highest_rating, total_battles, updated_at
		 FROM pvp_stats ORDER BY rating DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var stats []*PvPStats
	for rows.Next() {
		s := &PvPStats{}
		var updatedAt string
		if err := rows.Scan(&s.TrainerID, &s.Wins, &s.Losses, &s.Draws,
			&s.Rating, &s.HighestRating, &s.TotalBattles, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan pvp stats: %w", err)
		}
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		stats = append(stats, s)
	}
	return stats, nil
}

