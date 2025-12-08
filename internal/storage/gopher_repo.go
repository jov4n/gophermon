package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Gopher struct {
	ID              string
	TrainerID       *string
	Name            string
	Level           int
	XP              int
	CurrentHP       int
	MaxHP           int
	Attack          int
	Defense         int
	Speed           int
	Rarity          string
	ComplexityScore int
	SpeciesArchetype string
	EvolutionStage  int
	PrimaryType     string  // Primary type (Hacker, Tank, Speedy, Support, Mage)
	SecondaryType   string  // Secondary type for dual-type gophers (optional)
	SpritePath      string  // Deprecated: kept for backward compatibility, can be empty
	SpriteData      string  // Base64 encoded PNG image data
	GopherkonLayers []string // Will be stored as JSON
	IsInParty       bool
	PCSlot          *int
	CreatedAt       time.Time
}

type GopherRepo struct {
	db *DB
}

func NewGopherRepo(db *DB) *GopherRepo {
	return &GopherRepo{db: db}
}

func (r *GopherRepo) Create(g *Gopher) (*Gopher, error) {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}

	layersJSON, err := json.Marshal(g.GopherkonLayers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal layers: %w", err)
	}

	query := `INSERT INTO gophers (
		id, trainer_id, name, level, xp, current_hp, max_hp, 
		attack, defense, speed, rarity, complexity_score, 
		species_archetype, evolution_stage, primary_type, secondary_type,
		sprite_path, sprite_data, gopherkon_layers, is_in_party, pc_slot
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.Conn().Exec(query,
		g.ID, g.TrainerID, g.Name, g.Level, g.XP,
		g.CurrentHP, g.MaxHP, g.Attack, g.Defense, g.Speed,
		g.Rarity, g.ComplexityScore, g.SpeciesArchetype,
		g.EvolutionStage, g.PrimaryType, g.SecondaryType,
		g.SpritePath, g.SpriteData, string(layersJSON),
		g.IsInParty, g.PCSlot,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create gopher: %w", err)
	}

	return r.GetByID(g.ID)
}

func (r *GopherRepo) GetByID(id string) (*Gopher, error) {
	query := `SELECT id, trainer_id, name, level, xp, current_hp, max_hp,
	          attack, defense, speed, rarity, complexity_score,
	          species_archetype, evolution_stage, primary_type, secondary_type,
	          sprite_path, sprite_data, gopherkon_layers, is_in_party, pc_slot, created_at
	          FROM gophers WHERE id = ?`

	var g Gopher
	var trainerID sql.NullString
	var pcSlot sql.NullInt64
	var spritePath sql.NullString
	var spriteData sql.NullString
	var primaryType sql.NullString
	var secondaryType sql.NullString
	var layersJSON string
	var createdAt string

	err := r.db.Conn().QueryRow(query, id).Scan(
		&g.ID, &trainerID, &g.Name, &g.Level, &g.XP,
		&g.CurrentHP, &g.MaxHP, &g.Attack, &g.Defense, &g.Speed,
		&g.Rarity, &g.ComplexityScore, &g.SpeciesArchetype,
		&g.EvolutionStage, &primaryType, &secondaryType,
		&spritePath, &spriteData, &layersJSON,
		&g.IsInParty, &pcSlot, &createdAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get gopher: %w", err)
	}

	if trainerID.Valid {
		g.TrainerID = &trainerID.String
	}
	if pcSlot.Valid {
		slot := int(pcSlot.Int64)
		g.PCSlot = &slot
	}
	if spritePath.Valid {
		g.SpritePath = spritePath.String
	}
	if spriteData.Valid {
		g.SpriteData = spriteData.String
	}
	if primaryType.Valid {
		g.PrimaryType = primaryType.String
	}
	if secondaryType.Valid {
		g.SecondaryType = secondaryType.String
	}

	if err := json.Unmarshal([]byte(layersJSON), &g.GopherkonLayers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal layers: %w", err)
	}

	g.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &g, nil
}

func (r *GopherRepo) GetByTrainerID(trainerID string) ([]*Gopher, error) {
	query := `SELECT id, trainer_id, name, level, xp, current_hp, max_hp,
	          attack, defense, speed, rarity, complexity_score,
	          species_archetype, evolution_stage, primary_type, secondary_type,
	          sprite_path, sprite_data, gopherkon_layers, is_in_party, pc_slot, created_at
	          FROM gophers WHERE trainer_id = ? ORDER BY is_in_party DESC, created_at ASC`

	rows, err := r.db.Conn().Query(query, trainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query gophers: %w", err)
	}
	defer rows.Close()

	var gophers []*Gopher
	for rows.Next() {
		g, err := r.scanGopherRow(rows)
		if err != nil {
			return nil, err
		}
		gophers = append(gophers, g)
	}

	return gophers, nil
}

func (r *GopherRepo) GetParty(trainerID string) ([]*Gopher, error) {
	query := `SELECT id, trainer_id, name, level, xp, current_hp, max_hp,
	          attack, defense, speed, rarity, complexity_score,
	          species_archetype, evolution_stage, primary_type, secondary_type,
	          sprite_path, sprite_data, gopherkon_layers, is_in_party, pc_slot, created_at
	          FROM gophers WHERE trainer_id = ? AND is_in_party = TRUE
	          ORDER BY created_at ASC LIMIT 6`

	rows, err := r.db.Conn().Query(query, trainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query party: %w", err)
	}
	defer rows.Close()

	var gophers []*Gopher
	for rows.Next() {
		g, err := r.scanGopherRow(rows)
		if err != nil {
			return nil, err
		}
		gophers = append(gophers, g)
	}

	return gophers, nil
}

func (r *GopherRepo) GetPC(trainerID string, limit, offset int) ([]*Gopher, error) {
	query := `SELECT id, trainer_id, name, level, xp, current_hp, max_hp,
	          attack, defense, speed, rarity, complexity_score,
	          species_archetype, evolution_stage, primary_type, secondary_type,
	          sprite_path, sprite_data, gopherkon_layers, is_in_party, pc_slot, created_at
	          FROM gophers WHERE trainer_id = ? AND is_in_party = FALSE
	          ORDER BY pc_slot ASC LIMIT ? OFFSET ?`

	rows, err := r.db.Conn().Query(query, trainerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query PC: %w", err)
	}
	defer rows.Close()

	var gophers []*Gopher
	for rows.Next() {
		g, err := r.scanGopherRow(rows)
		if err != nil {
			return nil, err
		}
		gophers = append(gophers, g)
	}

	return gophers, nil
}

func (r *GopherRepo) Update(g *Gopher) error {
	layersJSON, err := json.Marshal(g.GopherkonLayers)
	if err != nil {
		return fmt.Errorf("failed to marshal layers: %w", err)
	}

	query := `UPDATE gophers SET
		trainer_id = ?, name = ?, level = ?, xp = ?, current_hp = ?, max_hp = ?,
		attack = ?, defense = ?, speed = ?, rarity = ?,
		complexity_score = ?, species_archetype = ?,
		evolution_stage = ?, primary_type = ?, secondary_type = ?,
		sprite_path = ?, sprite_data = ?, gopherkon_layers = ?,
		is_in_party = ?, pc_slot = ?
		WHERE id = ?`

	_, err = r.db.Conn().Exec(query,
		g.TrainerID, g.Name, g.Level, g.XP, g.CurrentHP, g.MaxHP,
		g.Attack, g.Defense, g.Speed, g.Rarity,
		g.ComplexityScore, g.SpeciesArchetype,
		g.EvolutionStage, g.PrimaryType, g.SecondaryType,
		g.SpritePath, g.SpriteData, string(layersJSON),
		g.IsInParty, g.PCSlot, g.ID,
	)

	return err
}

func (r *GopherRepo) Delete(id string) error {
	query := `DELETE FROM gophers WHERE id = ?`
	_, err := r.db.Conn().Exec(query, id)
	return err
}

func (r *GopherRepo) CountPC(trainerID string) (int, error) {
	query := `SELECT COUNT(*) FROM gophers WHERE trainer_id = ? AND is_in_party = FALSE`
	var count int
	err := r.db.Conn().QueryRow(query, trainerID).Scan(&count)
	return count, err
}

// scanGopherRow is a helper to scan a gopher row from a query result
func (r *GopherRepo) scanGopherRow(rows *sql.Rows) (*Gopher, error) {
	var g Gopher
	var trainerID sql.NullString
	var pcSlot sql.NullInt64
	var spritePath sql.NullString
	var spriteData sql.NullString
	var primaryType sql.NullString
	var secondaryType sql.NullString
	var layersJSON string
	var createdAt string

	err := rows.Scan(
		&g.ID, &trainerID, &g.Name, &g.Level, &g.XP,
		&g.CurrentHP, &g.MaxHP, &g.Attack, &g.Defense, &g.Speed,
		&g.Rarity, &g.ComplexityScore, &g.SpeciesArchetype,
		&g.EvolutionStage, &primaryType, &secondaryType,
		&spritePath, &spriteData, &layersJSON,
		&g.IsInParty, &pcSlot, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan gopher: %w", err)
	}

	if trainerID.Valid {
		g.TrainerID = &trainerID.String
	}
	if pcSlot.Valid {
		slot := int(pcSlot.Int64)
		g.PCSlot = &slot
	}
	if spritePath.Valid {
		g.SpritePath = spritePath.String
	}
	if spriteData.Valid {
		g.SpriteData = spriteData.String
	}
	if primaryType.Valid {
		g.PrimaryType = primaryType.String
	}
	if secondaryType.Valid {
		g.SecondaryType = secondaryType.String
	}

	if err := json.Unmarshal([]byte(layersJSON), &g.GopherkonLayers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal layers: %w", err)
	}

	g.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &g, nil
}

