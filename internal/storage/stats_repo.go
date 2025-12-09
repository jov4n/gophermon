package storage

import (
	"database/sql"
	"fmt"
	"time"
)

type TrainerStats struct {
	TrainerID      string
	TotalBattles   int
	BattlesWon     int
	BattlesLost    int
	GophersCaught  int
	ShinyCount     int
	Evolutions     int
	TotalXPEarned  int
	FavoriteArchetype *string
	MostUsedGopherID  *string
	UpdatedAt      time.Time
}

type StatsRepo struct {
	db *DB
}

func NewStatsRepo(db *DB) *StatsRepo {
	return &StatsRepo{db: db}
}

func (r *StatsRepo) GetOrCreate(trainerID string) (*TrainerStats, error) {
	query := `SELECT trainer_id, total_battles, battles_won, battles_lost, gophers_caught, 
	          shiny_count, evolutions, total_xp_earned, favorite_archetype, most_used_gopher_id, updated_at
	          FROM trainer_stats WHERE trainer_id = ?`
	
	var stats TrainerStats
	var favoriteArchetype, mostUsedGopherID sql.NullString
	var updatedAt string
	
	err := r.db.Conn().QueryRow(query, trainerID).Scan(
		&stats.TrainerID, &stats.TotalBattles, &stats.BattlesWon, &stats.BattlesLost,
		&stats.GophersCaught, &stats.ShinyCount, &stats.Evolutions, &stats.TotalXPEarned,
		&favoriteArchetype, &mostUsedGopherID, &updatedAt,
	)
	
	if err == sql.ErrNoRows {
		// Create new stats
		_, err = r.db.Conn().Exec(
			`INSERT INTO trainer_stats (trainer_id) VALUES (?)`,
			trainerID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stats: %w", err)
		}
		return r.GetOrCreate(trainerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	if favoriteArchetype.Valid {
		stats.FavoriteArchetype = &favoriteArchetype.String
	}
	if mostUsedGopherID.Valid {
		stats.MostUsedGopherID = &mostUsedGopherID.String
	}
	stats.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return &stats, nil
}

func (r *StatsRepo) IncrementBattles(trainerID string, won bool) error {
	var wins, losses int
	if won {
		wins = 1
	} else {
		losses = 1
	}

	_, err := r.db.Conn().Exec(
		`UPDATE trainer_stats SET total_battles = total_battles + 1, 
		 battles_won = battles_won + ?, battles_lost = battles_lost + ?,
		 updated_at = CURRENT_TIMESTAMP WHERE trainer_id = ?`,
		wins, losses, trainerID,
	)
	return err
}

func (r *StatsRepo) IncrementGophersCaught(trainerID string, isShiny bool) error {
	var shinyCount int
	if isShiny {
		shinyCount = 1
	}

	_, err := r.db.Conn().Exec(
		`UPDATE trainer_stats SET gophers_caught = gophers_caught + 1,
		 shiny_count = shiny_count + ?, updated_at = CURRENT_TIMESTAMP WHERE trainer_id = ?`,
		shinyCount, trainerID,
	)
	return err
}

func (r *StatsRepo) IncrementEvolutions(trainerID string) error {
	_, err := r.db.Conn().Exec(
		`UPDATE trainer_stats SET evolutions = evolutions + 1, updated_at = CURRENT_TIMESTAMP 
		 WHERE trainer_id = ?`,
		trainerID,
	)
	return err
}

func (r *StatsRepo) AddXP(trainerID string, xp int) error {
	_, err := r.db.Conn().Exec(
		`UPDATE trainer_stats SET total_xp_earned = total_xp_earned + ?, updated_at = CURRENT_TIMESTAMP 
		 WHERE trainer_id = ?`,
		xp, trainerID,
	)
	return err
}

func (r *StatsRepo) GetLeaderboard(statType string, limit int) ([]*TrainerStats, error) {
	var orderBy string
	switch statType {
	case "wins":
		orderBy = "battles_won DESC"
	case "shinies":
		orderBy = "shiny_count DESC"
	case "caught":
		orderBy = "gophers_caught DESC"
	case "xp":
		orderBy = "total_xp_earned DESC"
	default:
		orderBy = "battles_won DESC"
	}

	rows, err := r.db.Conn().Query(
		`SELECT trainer_id, total_battles, battles_won, battles_lost, gophers_caught,
		 shiny_count, evolutions, total_xp_earned, favorite_archetype, most_used_gopher_id, updated_at
		 FROM trainer_stats ORDER BY `+orderBy+` LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var stats []*TrainerStats
	for rows.Next() {
		s := &TrainerStats{}
		var favoriteArchetype, mostUsedGopherID sql.NullString
		var updatedAt string
		if err := rows.Scan(&s.TrainerID, &s.TotalBattles, &s.BattlesWon, &s.BattlesLost,
			&s.GophersCaught, &s.ShinyCount, &s.Evolutions, &s.TotalXPEarned,
			&favoriteArchetype, &mostUsedGopherID, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}
		if favoriteArchetype.Valid {
			s.FavoriteArchetype = &favoriteArchetype.String
		}
		if mostUsedGopherID.Valid {
			s.MostUsedGopherID = &mostUsedGopherID.String
		}
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
		stats = append(stats, s)
	}
	return stats, nil
}

