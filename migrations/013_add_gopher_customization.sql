-- Migration to add gopher customization features

ALTER TABLE gophers ADD COLUMN nickname TEXT;
ALTER TABLE gophers ADD COLUMN is_favorite BOOLEAN DEFAULT FALSE;
ALTER TABLE gophers ADD COLUMN tags TEXT DEFAULT '[]'; -- JSON array of tags

CREATE INDEX IF NOT EXISTS idx_gophers_is_favorite ON gophers(is_favorite);

