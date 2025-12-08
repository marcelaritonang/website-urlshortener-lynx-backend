CREATE TABLE IF NOT EXISTS urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,
    user_id UUID NOT NULL,
    long_url TEXT NOT NULL,
    short_code VARCHAR(10) NOT NULL UNIQUE,
    clicks INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMP DEFAULT NULL,
    CONSTRAINT fk_users_urls FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls(user_id);
CREATE INDEX IF NOT EXISTS idx_urls_deleted_at ON urls(deleted_at);