-- Migration to add Gopherdex (collection tracking)

CREATE TABLE IF NOT EXISTS gopherdex (
    trainer_id TEXT NOT NULL,
    gopher_name TEXT NOT NULL,
    archetype TEXT NOT NULL,
    rarity TEXT NOT NULL,
    first_encountered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    times_encountered INTEGER DEFAULT 1,
    times_caught INTEGER DEFAULT 0,
    owned BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (trainer_id, gopher_name, archetype, rarity),
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE
);

-- Create unique index for conflict resolution
CREATE UNIQUE INDEX IF NOT EXISTS idx_gopherdex_unique ON gopherdex(trainer_id, gopher_name, archetype, rarity);

CREATE INDEX IF NOT EXISTS idx_gopherdex_trainer_id ON gopherdex(trainer_id);
CREATE INDEX IF NOT EXISTS idx_gopherdex_owned ON gopherdex(owned);

