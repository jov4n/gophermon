-- Migration to add PvP battle system

ALTER TABLE battles ADD COLUMN opponent_trainer_id TEXT;
ALTER TABLE battles ADD COLUMN battle_rating INTEGER DEFAULT 1000;

CREATE TABLE IF NOT EXISTS pvp_stats (
    trainer_id TEXT PRIMARY KEY,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,
    rating INTEGER DEFAULT 1000,
    highest_rating INTEGER DEFAULT 1000,
    total_battles INTEGER DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pvp_stats_rating ON pvp_stats(rating);

