-- Migration to add shiny field for gophers
-- Shiny gophers are rare variants with inverted colors

ALTER TABLE gophers ADD COLUMN shiny BOOLEAN DEFAULT FALSE;

