-- Migration to add statistics tracking

CREATE TABLE IF NOT EXISTS trainer_stats (
    trainer_id TEXT PRIMARY KEY,
    total_battles INTEGER DEFAULT 0,
    battles_won INTEGER DEFAULT 0,
    battles_lost INTEGER DEFAULT 0,
    gophers_caught INTEGER DEFAULT 0,
    shiny_count INTEGER DEFAULT 0,
    evolutions INTEGER DEFAULT 0,
    total_xp_earned INTEGER DEFAULT 0,
    favorite_archetype TEXT,
    most_used_gopher_id TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE,
    FOREIGN KEY (most_used_gopher_id) REFERENCES gophers(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_trainer_stats_battles_won ON trainer_stats(battles_won);
CREATE INDEX IF NOT EXISTS idx_trainer_stats_shiny_count ON trainer_stats(shiny_count);

