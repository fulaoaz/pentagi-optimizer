-- +goose Up
-- +goose StatementBegin
-- Add cvss to the searchengine_type enum
CREATE TYPE SEARCHENGINE_TYPE_NEW AS ENUM (
  'google',
  'tavily',
  'firecrawl',
  'traversaal',
  'browser',
  'duckduckgo',
  'perplexity',
  'searxng',
  'sploitus',
  'eppss',
  'cvss'
);

-- Update the searchlogs table to use the new enum type
ALTER TABLE searchlogs
    ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_NEW USING engine::text::SEARCHENGINE_TYPE_NEW;

-- Drop the old type and rename the new one
DROP TYPE SEARCHENGINE_TYPE;
ALTER TYPE SEARCHENGINE_TYPE_NEW RENAME TO SEARCHENGINE_TYPE;

-- Ensure NOT NULL constraint is preserved
ALTER TABLE searchlogs
    ALTER COLUMN engine SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert the changes by removing cvss from the enum
CREATE TYPE SEARCHENGINE_TYPE_NEW AS ENUM (
  'google',
  'tavily',
  'firecrawl',
  'traversaal',
  'browser',
  'duckduckgo',
  'perplexity',
  'searxng',
  'sploitus',
  'eppss'
);

-- Remap any rows logged with the removed value so the narrowing cast below
-- cannot fail; 'cvss' scores vulnerabilities like 'eppss'.
UPDATE searchlogs
    SET engine = 'eppss'
    WHERE engine = 'cvss';

-- Update the searchlogs table to use the reverted enum type
ALTER TABLE searchlogs
    ALTER COLUMN engine TYPE SEARCHENGINE_TYPE_NEW USING engine::text::SEARCHENGINE_TYPE_NEW;

-- Drop the new type and rename the reverted one
DROP TYPE SEARCHENGINE_TYPE;
ALTER TYPE SEARCHENGINE_TYPE_NEW RENAME TO SEARCHENGINE_TYPE;

-- Ensure NOT NULL constraint is preserved
ALTER TABLE searchlogs
    ALTER COLUMN engine SET NOT NULL;
-- +goose StatementEnd
