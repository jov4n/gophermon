package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Quest struct {
	ID           string
	TrainerID    string
	QuestType    string
	QuestName    string
	Description  string
	TargetValue  int
	CurrentProgress int
	RewardCurrency int
	RewardXP     int
	Completed    bool
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type QuestRepo struct {
	db *DB
}

func NewQuestRepo(db *DB) *QuestRepo {
	return &QuestRepo{db: db}
}

func (r *QuestRepo) Create(quest *Quest) error {
	if quest.ID == "" {
		quest.ID = uuid.New().String()
	}
	_, err := r.db.Conn().Exec(
		`INSERT INTO quests (id, trainer_id, quest_type, quest_name, description, target_value, 
		 current_progress, reward_currency, reward_xp, completed, expires_at) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		quest.ID, quest.TrainerID, quest.QuestType, quest.QuestName, quest.Description,
		quest.TargetValue, quest.CurrentProgress, quest.RewardCurrency, quest.RewardXP,
		quest.Completed, quest.ExpiresAt,
	)
	return err
}

func (r *QuestRepo) UpdateProgress(questID string, progress int) error {
	_, err := r.db.Conn().Exec(
		"UPDATE quests SET current_progress = current_progress + ? WHERE id = ?",
		progress, questID,
	)
	return err
}

func (r *QuestRepo) Complete(questID string) error {
	_, err := r.db.Conn().Exec(
		"UPDATE quests SET completed = TRUE WHERE id = ?",
		questID,
	)
	return err
}

func (r *QuestRepo) GetActiveQuests(trainerID string) ([]*Quest, error) {
	rows, err := r.db.Conn().Query(
		`SELECT id, trainer_id, quest_type, quest_name, description, target_value, 
		 current_progress, reward_currency, reward_xp, completed, expires_at, created_at
		 FROM quests WHERE trainer_id = ? AND completed = FALSE AND expires_at > CURRENT_TIMESTAMP`,
		trainerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query quests: %w", err)
	}
	defer rows.Close()

	var quests []*Quest
	for rows.Next() {
		quest := &Quest{}
		var expiresAt, createdAt string
		if err := rows.Scan(&quest.ID, &quest.TrainerID, &quest.QuestType, &quest.QuestName,
			&quest.Description, &quest.TargetValue, &quest.CurrentProgress, &quest.RewardCurrency,
			&quest.RewardXP, &quest.Completed, &expiresAt, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan quest: %w", err)
		}
		quest.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)
		quest.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		quests = append(quests, quest)
	}
	return quests, nil
}

func (r *QuestRepo) GetQuestByType(trainerID, questType, questName string) (*Quest, error) {
	query := `SELECT id, trainer_id, quest_type, quest_name, description, target_value, 
	          current_progress, reward_currency, reward_xp, completed, expires_at, created_at
	          FROM quests WHERE trainer_id = ? AND quest_type = ? AND quest_name = ? 
	          AND completed = FALSE AND expires_at > CURRENT_TIMESTAMP`
	
	quest := &Quest{}
	var expiresAt, createdAt string
	err := r.db.Conn().QueryRow(query, trainerID, questType, questName).Scan(
		&quest.ID, &quest.TrainerID, &quest.QuestType, &quest.QuestName,
		&quest.Description, &quest.TargetValue, &quest.CurrentProgress, &quest.RewardCurrency,
		&quest.RewardXP, &quest.Completed, &expiresAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}
	quest.ExpiresAt, _ = time.Parse("2006-01-02 15:04:05", expiresAt)
	quest.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return quest, nil
}

func (r *QuestRepo) CleanupExpired() error {
	_, err := r.db.Conn().Exec(
		"DELETE FROM quests WHERE expires_at < CURRENT_TIMESTAMP AND completed = FALSE",
	)
	return err
}

