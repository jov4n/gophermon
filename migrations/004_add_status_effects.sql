-- Migration to add status_effects field for gophers
-- Status effects are stored as JSON array

ALTER TABLE gophers ADD COLUMN status_effects TEXT DEFAULT '[]';

