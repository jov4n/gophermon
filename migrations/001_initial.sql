-- Trainers table
CREATE TABLE IF NOT EXISTS trainers (
    id TEXT PRIMARY KEY,
    discord_id TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    active_party_slots INTEGER DEFAULT 0
);

-- Gophers table
CREATE TABLE IF NOT EXISTS gophers (
    id TEXT PRIMARY KEY,
    trainer_id TEXT,
    name TEXT NOT NULL,
    level INTEGER DEFAULT 1,
    xp INTEGER DEFAULT 0,
    current_hp INTEGER NOT NULL,
    max_hp INTEGER NOT NULL,
    attack INTEGER NOT NULL,
    defense INTEGER NOT NULL,
    speed INTEGER NOT NULL,
    rarity TEXT NOT NULL CHECK(rarity IN ('COMMON', 'UNCOMMON', 'RARE', 'EPIC', 'LEGENDARY')),
    complexity_score INTEGER NOT NULL,
    species_archetype TEXT NOT NULL,
    evolution_stage INTEGER DEFAULT 0,
    sprite_path TEXT,
    gopherkon_layers TEXT NOT NULL, -- JSON string
    is_in_party BOOLEAN DEFAULT FALSE,
    pc_slot INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE SET NULL
);

-- Abilities table
CREATE TABLE IF NOT EXISTS abilities (
    id TEXT PRIMARY KEY,
    gopher_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    power INTEGER NOT NULL,
    cost INTEGER NOT NULL,
    targeting TEXT NOT NULL CHECK(targeting IN ('SELF', 'ENEMY', 'BOTH')),
    effect_json TEXT, -- JSON string for status effects, buff/debuff rules
    FOREIGN KEY (gopher_id) REFERENCES gophers(id) ON DELETE CASCADE
);

-- Battles table
CREATE TABLE IF NOT EXISTS battles (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    trainer_id TEXT NOT NULL,
    opponent_type TEXT NOT NULL CHECK(opponent_type IN ('WILD', 'TRAINER', 'PVE_BOSS')),
    gopher_id_player TEXT,
    gopher_id_enemy TEXT,
    turn_owner TEXT NOT NULL CHECK(turn_owner IN ('PLAYER', 'ENEMY')),
    state TEXT NOT NULL CHECK(state IN ('ACTIVE', 'WON', 'LOST', 'ESCAPED')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (trainer_id) REFERENCES trainers(id) ON DELETE CASCADE,
    FOREIGN KEY (gopher_id_player) REFERENCES gophers(id) ON DELETE SET NULL,
    FOREIGN KEY (gopher_id_enemy) REFERENCES gophers(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_gophers_trainer_id ON gophers(trainer_id);
CREATE INDEX IF NOT EXISTS idx_gophers_is_in_party ON gophers(is_in_party);
CREATE INDEX IF NOT EXISTS idx_abilities_gopher_id ON abilities(gopher_id);
CREATE INDEX IF NOT EXISTS idx_battles_trainer_id ON battles(trainer_id);
CREATE INDEX IF NOT EXISTS idx_battles_state ON battles(state);
CREATE INDEX IF NOT EXISTS idx_trainers_discord_id ON trainers(discord_id);

