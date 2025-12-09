-- Migration to add trading system

CREATE TABLE IF NOT EXISTS trades (
    id TEXT PRIMARY KEY,
    trainer1_id TEXT NOT NULL,
    trainer2_id TEXT NOT NULL,
    gopher1_id TEXT,
    gopher2_id TEXT,
    currency1 INTEGER DEFAULT 0,
    currency2 INTEGER DEFAULT 0,
    status TEXT NOT NULL CHECK(status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'CANCELLED')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (trainer1_id) REFERENCES trainers(id) ON DELETE CASCADE,
    FOREIGN KEY (trainer2_id) REFERENCES trainers(id) ON DELETE CASCADE,
    FOREIGN KEY (gopher1_id) REFERENCES gophers(id) ON DELETE SET NULL,
    FOREIGN KEY (gopher2_id) REFERENCES gophers(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_trades_trainer1 ON trades(trainer1_id);
CREATE INDEX IF NOT EXISTS idx_trades_trainer2 ON trades(trainer2_id);
CREATE INDEX IF NOT EXISTS idx_trades_status ON trades(status);

