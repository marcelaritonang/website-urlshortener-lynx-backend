CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,
    user_id INTEGER NOT NULL,
    original_url TEXT NOT NULL,
    short_url VARCHAR(255) NOT NULL UNIQUE,
    visits INTEGER NOT NULL DEFAULT 0,
    expires_at DATETIME DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_urls_short_url ON urls(short_url);
CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE INDEX idx_urls_deleted_at ON urls(deleted_at);