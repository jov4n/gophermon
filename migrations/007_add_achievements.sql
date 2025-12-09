-- Migration to add achievement system

CREATE TABLE IF NOT EXISTS achievements (
    id TEXT PRIMARY KEY,
    trainer_id TEXT NOT NULL,
    achievement_type TEXT NOT NULL,
    progress INTEGER DEFAULT 0,
    completed BOOLEAN DEFAULT FALSE,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE,
    UNIQUE(trainer_id, achievement_type)
);

CREATE INDEX IF NOT EXISTS idx_achievements_trainer_id ON achievements(trainer_id);
CREATE INDEX IF NOT EXISTS idx_achievements_type ON achievements(achievement_type);
CREATE INDEX IF NOT EXISTS idx_achievements_completed ON achievements(completed);

