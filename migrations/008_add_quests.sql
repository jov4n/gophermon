-- Migration to add daily/weekly quest system

CREATE TABLE IF NOT EXISTS quests (
    id TEXT PRIMARY KEY,
    trainer_id TEXT NOT NULL,
    quest_type TEXT NOT NULL CHECK(quest_type IN ('DAILY', 'WEEKLY')),
    quest_name TEXT NOT NULL,
    description TEXT NOT NULL,
    target_value INTEGER NOT NULL,
    current_progress INTEGER DEFAULT 0,
    reward_currency INTEGER DEFAULT 0,
    reward_xp INTEGER DEFAULT 0,
    completed BOOLEAN DEFAULT FALSE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_quests_trainer_id ON quests(trainer_id);
CREATE INDEX IF NOT EXISTS idx_quests_type ON quests(quest_type);
CREATE INDEX IF NOT EXISTS idx_quests_expires_at ON quests(expires_at);

