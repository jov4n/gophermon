-- Migration to add economy system (currency and items)
-- Currency: GoCoins (gopher coins)

ALTER TABLE trainers ADD COLUMN currency INTEGER DEFAULT 100;

-- Items table
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    trainer_id TEXT NOT NULL,
    item_type TEXT NOT NULL CHECK(item_type IN ('POTION', 'REVIVE', 'XP_BOOSTER', 'EVOLUTION_STONE', 'SHINY_CHARM')),
    quantity INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_items_trainer_id ON items(trainer_id);
CREATE INDEX IF NOT EXISTS idx_items_type ON items(item_type);

