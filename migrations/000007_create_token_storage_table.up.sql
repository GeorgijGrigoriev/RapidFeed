CREATE TABLE IF NOT EXISTS token_storage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    token TEXT NOT NULL,
    expires_at INTEGER,
    permissions INTEGER
);
