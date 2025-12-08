-- Migration to add sprite_data column for base64 image storage
-- Make sprite_path nullable to support both storage methods during transition

-- Note: SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN
-- The migration system will handle checking if this migration has been run
-- If you need to manually run this, check if the column exists first

-- Add sprite_data column
ALTER TABLE gophers ADD COLUMN sprite_data TEXT;

-- Note: sprite_path is kept for backward compatibility but can be NULL
-- New gophers will use sprite_data (base64), old ones may still have sprite_path

