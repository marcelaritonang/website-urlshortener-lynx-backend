-- Make user_id nullable (for anonymous URLs)
ALTER TABLE urls ALTER COLUMN user_id DROP NOT NULL;

-- Add new columns
ALTER TABLE urls ADD COLUMN IF NOT EXISTS short_code VARCHAR(10);
ALTER TABLE urls ADD COLUMN IF NOT EXISTS is_anonymous BOOLEAN DEFAULT FALSE;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_is_anonymous ON urls(is_anonymous);
CREATE INDEX IF NOT EXISTS idx_urls_deleted_at ON urls(deleted_at);

-- Migrate existing data: Extract short_code from short_url
-- Example: "http://localhost:8080/urls/aN63Mw" → "aN63Mw"
UPDATE urls 
SET short_code = SUBSTRING(short_url FROM '[^/]+$')
WHERE short_code IS NULL;

-- Make short_code required after migration
ALTER TABLE urls ALTER COLUMN short_code SET NOT NULL;