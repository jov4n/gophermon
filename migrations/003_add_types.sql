-- Migration to add type fields for type effectiveness system
-- Gophers can have a primary type and optional secondary type (for dual types)

ALTER TABLE gophers ADD COLUMN primary_type TEXT;
ALTER TABLE gophers ADD COLUMN secondary_type TEXT;

-- Set default types based on archetype for existing gophers
-- This will be handled by application code, but we set defaults here
UPDATE gophers SET primary_type = species_archetype WHERE primary_type IS NULL;

