package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Trade struct {
	ID          string
	Trainer1ID  string
	Trainer2ID  string
	Gopher1ID   *string
	Gopher2ID   *string
	Currency1  int
	Currency2  int
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

type TradeRepo struct {
	db *DB
}

func NewTradeRepo(db *DB) *TradeRepo {
	return &TradeRepo{db: db}
}

func (r *TradeRepo) Create(trade *Trade) error {
	if trade.ID == "" {
		trade.ID = uuid.New().String()
	}
	_, err := r.db.Conn().Exec(
		`INSERT INTO trades (id, trainer1_id, trainer2_id, gopher1_id, gopher2_id, 
		 currency1, currency2, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		trade.ID, trade.Trainer1ID, trade.Trainer2ID, trade.Gopher1ID, trade.Gopher2ID,
		trade.Currency1, trade.Currency2, trade.Status,
	)
	return err
}

func (r *TradeRepo) GetByID(tradeID string) (*Trade, error) {
	query := `SELECT id, trainer1_id, trainer2_id, gopher1_id, gopher2_id, 
	          currency1, currency2, status, created_at, completed_at
	          FROM trades WHERE id = ?`
	
	trade := &Trade{}
	var gopher1ID, gopher2ID sql.NullString
	var completedAt sql.NullString
	var createdAt string
	
	err := r.db.Conn().QueryRow(query, tradeID).Scan(
		&trade.ID, &trade.Trainer1ID, &trade.Trainer2ID, &gopher1ID, &gopher2ID,
		&trade.Currency1, &trade.Currency2, &trade.Status, &createdAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trade: %w", err)
	}

	if gopher1ID.Valid {
		trade.Gopher1ID = &gopher1ID.String
	}
	if gopher2ID.Valid {
		trade.Gopher2ID = &gopher2ID.String
	}
	if completedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
		trade.CompletedAt = &t
	}
	trade.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return trade, nil
}

func (r *TradeRepo) GetPendingTrades(trainerID string) ([]*Trade, error) {
	rows, err := r.db.Conn().Query(
		`SELECT id, trainer1_id, trainer2_id, gopher1_id, gopher2_id, 
		 currency1, currency2, status, created_at, completed_at
		 FROM trades WHERE (trainer1_id = ? OR trainer2_id = ?) AND status = 'PENDING'`,
		trainerID, trainerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query trades: %w", err)
	}
	defer rows.Close()

	var trades []*Trade
	for rows.Next() {
		trade := &Trade{}
		var gopher1ID, gopher2ID sql.NullString
		var completedAt sql.NullString
		var createdAt string
		if err := rows.Scan(&trade.ID, &trade.Trainer1ID, &trade.Trainer2ID, &gopher1ID, &gopher2ID,
			&trade.Currency1, &trade.Currency2, &trade.Status, &createdAt, &completedAt); err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		if gopher1ID.Valid {
			trade.Gopher1ID = &gopher1ID.String
		}
		if gopher2ID.Valid {
			trade.Gopher2ID = &gopher2ID.String
		}
		if completedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", completedAt.String)
			trade.CompletedAt = &t
		}
		trade.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		trades = append(trades, trade)
	}
	return trades, nil
}

func (r *TradeRepo) UpdateStatus(tradeID, status string) error {
	query := `UPDATE trades SET status = ?, completed_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Conn().Exec(query, status, tradeID)
	return err
}

