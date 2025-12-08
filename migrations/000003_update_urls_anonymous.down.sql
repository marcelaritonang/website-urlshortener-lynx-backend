-- Remove new columns
ALTER TABLE urls DROP COLUMN IF EXISTS short_code;
ALTER TABLE urls DROP COLUMN IF EXISTS is_anonymous;
ALTER TABLE urls DROP COLUMN IF EXISTS expires_at;
ALTER TABLE urls DROP COLUMN IF EXISTS deleted_at;

-- Make user_id required again
ALTER TABLE urls ALTER COLUMN user_id SET NOT NULL;

-- Drop indexes
DROP INDEX IF EXISTS idx_urls_short_code;
DROP INDEX IF EXISTS idx_urls_is_anonymous;
DROP INDEX IF EXISTS idx_urls_deleted_at;