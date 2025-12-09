package storage

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type Item struct {
	ID        string
	TrainerID string
	ItemType  string
	Quantity  int
}

type ItemRepo struct {
	db *DB
}

func NewItemRepo(db *DB) *ItemRepo {
	return &ItemRepo{db: db}
}

func (r *ItemRepo) AddItem(trainerID, itemType string, quantity int) error {
	// Check if item already exists
	var existingID string
	var existingQty int
	err := r.db.Conn().QueryRow(
		"SELECT id, quantity FROM items WHERE trainer_id = ? AND item_type = ?",
		trainerID, itemType,
	).Scan(&existingID, &existingQty)

	if err == sql.ErrNoRows {
		// Create new item
		id := uuid.New().String()
		_, err = r.db.Conn().Exec(
			"INSERT INTO items (id, trainer_id, item_type, quantity) VALUES (?, ?, ?, ?)",
			id, trainerID, itemType, quantity,
		)
		return err
	} else if err != nil {
		return fmt.Errorf("failed to check existing item: %w", err)
	}

	// Update existing item
	_, err = r.db.Conn().Exec(
		"UPDATE items SET quantity = quantity + ? WHERE id = ?",
		quantity, existingID,
	)
	return err
}

func (r *ItemRepo) UseItem(trainerID, itemType string, quantity int) error {
	var currentQty int
	err := r.db.Conn().QueryRow(
		"SELECT quantity FROM items WHERE trainer_id = ? AND item_type = ?",
		trainerID, itemType,
	).Scan(&currentQty)

	if err == sql.ErrNoRows {
		return fmt.Errorf("item not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	if currentQty < quantity {
		return fmt.Errorf("insufficient quantity")
	}

	newQty := currentQty - quantity
	if newQty <= 0 {
		// Delete item
		_, err = r.db.Conn().Exec(
			"DELETE FROM items WHERE trainer_id = ? AND item_type = ?",
			trainerID, itemType,
		)
	} else {
		// Update quantity
		_, err = r.db.Conn().Exec(
			"UPDATE items SET quantity = ? WHERE trainer_id = ? AND item_type = ?",
			newQty, trainerID, itemType,
		)
	}
	return err
}

func (r *ItemRepo) GetItems(trainerID string) ([]*Item, error) {
	rows, err := r.db.Conn().Query(
		"SELECT id, trainer_id, item_type, quantity FROM items WHERE trainer_id = ?",
		trainerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item := &Item{}
		if err := rows.Scan(&item.ID, &item.TrainerID, &item.ItemType, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *ItemRepo) GetItemQuantity(trainerID, itemType string) (int, error) {
	var quantity int
	err := r.db.Conn().QueryRow(
		"SELECT COALESCE(SUM(quantity), 0) FROM items WHERE trainer_id = ? AND item_type = ?",
		trainerID, itemType,
	).Scan(&quantity)
	if err != nil {
		return 0, err
	}
	return quantity, nil
}

