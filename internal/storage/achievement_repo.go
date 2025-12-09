package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Achievement struct {
	ID            string
	TrainerID     string
	AchievementType string
	Progress      int
	Completed     bool
	CompletedAt   *time.Time
	CreatedAt     time.Time
}

type AchievementRepo struct {
	db *DB
}

func NewAchievementRepo(db *DB) *AchievementRepo {
	return &AchievementRepo{db: db}
}

func (r *AchievementRepo) GetOrCreate(trainerID, achievementType string) (*Achievement, error) {
	query := `SELECT id, trainer_id, achievement_type, progress, completed, completed_at, created_at
	          FROM achievements WHERE trainer_id = ? AND achievement_type = ?`
	
	var ach Achievement
	var completedAt sql.NullString
	var createdAt string
	
	err := r.db.Conn().QueryRow(query, trainerID, achievementType).Scan(
		&ach.ID, &ach.TrainerID, &ach.AchievementType, &ach.Progress, &ach.Completed,
		&completedAt, &createdAt,
	)
	
	if err == sql.ErrNoRows {
		// Create new achievement
		id := uuid.New().String()
		_, err = r.db.Conn().Exec(
			"INSERT INTO achievements (id, trainer_id, achievement_type, progress, completed) VALUES (?, ?, ?, 0, FALSE)",
			id, trainerID, achievementType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create achievement: %w", err)
		}
		return r.GetOrCreate(trainerID, achievementType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}

	if completedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
		ach.CompletedAt = &t
	}
	ach.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &ach, nil
}

func (r *AchievementRepo) UpdateProgress(trainerID, achievementType string, progress int) error {
	ach, err := r.GetOrCreate(trainerID, achievementType)
	if err != nil {
		return err
	}

	newProgress := ach.Progress + progress
	_, err = r.db.Conn().Exec(
		"UPDATE achievements SET progress = ? WHERE id = ?",
		newProgress, ach.ID,
	)
	return err
}

func (r *AchievementRepo) Complete(trainerID, achievementType string) error {
	ach, err := r.GetOrCreate(trainerID, achievementType)
	if err != nil {
		return err
	}

	if ach.Completed {
		return nil // Already completed
	}

	_, err = r.db.Conn().Exec(
		"UPDATE achievements SET completed = TRUE, completed_at = CURRENT_TIMESTAMP WHERE id = ?",
		ach.ID,
	)
	return err
}

func (r *AchievementRepo) GetAchievements(trainerID string) ([]*Achievement, error) {
	rows, err := r.db.Conn().Query(
		"SELECT id, trainer_id, achievement_type, progress, completed, completed_at, created_at FROM achievements WHERE trainer_id = ?",
		trainerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query achievements: %w", err)
	}
	defer rows.Close()

	var achievements []*Achievement
	for rows.Next() {
		ach := &Achievement{}
		var completedAt sql.NullString
		var createdAt string
		if err := rows.Scan(&ach.ID, &ach.TrainerID, &ach.AchievementType, &ach.Progress, &ach.Completed, &completedAt, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		if completedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			ach.CompletedAt = &t
		}
		ach.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		achievements = append(achievements, ach)
	}
	return achievements, nil
}

