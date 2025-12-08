package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Battle struct {
	ID           string
	ChannelID    string
	MessageID    string
	TrainerID    string
	OpponentType string
	GopherIDPlayer *string
	GopherIDEnemy  *string
	TurnOwner    string
	State        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type BattleRepo struct {
	db *DB
}

func NewBattleRepo(db *DB) *BattleRepo {
	return &BattleRepo{db: db}
}

func (r *BattleRepo) Create(b *Battle) (*Battle, error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}

	query := `INSERT INTO battles (
		id, channel_id, message_id, trainer_id, opponent_type,
		gopher_id_player, gopher_id_enemy, turn_owner, state
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Conn().Exec(query,
		b.ID, b.ChannelID, b.MessageID, b.TrainerID, b.OpponentType,
		b.GopherIDPlayer, b.GopherIDEnemy, b.TurnOwner, b.State,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create battle: %w", err)
	}

	return r.GetByID(b.ID)
}

func (r *BattleRepo) GetByID(id string) (*Battle, error) {
	query := `SELECT id, channel_id, message_id, trainer_id, opponent_type,
	          gopher_id_player, gopher_id_enemy, turn_owner, state,
	          created_at, updated_at
	          FROM battles WHERE id = ?`

	var b Battle
	var playerID sql.NullString
	var enemyID sql.NullString
	var createdAt, updatedAt string

	err := r.db.Conn().QueryRow(query, id).Scan(
		&b.ID, &b.ChannelID, &b.MessageID, &b.TrainerID, &b.OpponentType,
		&playerID, &enemyID, &b.TurnOwner, &b.State,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get battle: %w", err)
	}

	if playerID.Valid {
		b.GopherIDPlayer = &playerID.String
	}
	if enemyID.Valid {
		b.GopherIDEnemy = &enemyID.String
	}

	b.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	b.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &b, nil
}

func (r *BattleRepo) GetByMessageID(channelID, messageID string) (*Battle, error) {
	query := `SELECT id, channel_id, message_id, trainer_id, opponent_type,
	          gopher_id_player, gopher_id_enemy, turn_owner, state,
	          created_at, updated_at
	          FROM battles WHERE channel_id = ? AND message_id = ? AND state = 'ACTIVE'`

	var b Battle
	var playerID sql.NullString
	var enemyID sql.NullString
	var createdAt, updatedAt string

	err := r.db.Conn().QueryRow(query, channelID, messageID).Scan(
		&b.ID, &b.ChannelID, &b.MessageID, &b.TrainerID, &b.OpponentType,
		&playerID, &enemyID, &b.TurnOwner, &b.State,
		&createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get battle: %w", err)
	}

	if playerID.Valid {
		b.GopherIDPlayer = &playerID.String
	}
	if enemyID.Valid {
		b.GopherIDEnemy = &enemyID.String
	}

	b.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	b.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &b, nil
}

func (r *BattleRepo) Update(b *Battle) error {
	query := `UPDATE battles SET
		gopher_id_player = ?, gopher_id_enemy = ?,
		turn_owner = ?, state = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := r.db.Conn().Exec(query,
		b.GopherIDPlayer, b.GopherIDEnemy,
		b.TurnOwner, b.State, b.ID,
	)

	return err
}

func (r *BattleRepo) Delete(id string) error {
	query := `DELETE FROM battles WHERE id = ?`
	_, err := r.db.Conn().Exec(query, id)
	return err
}

