CREATE TABLE IF NOT EXISTS user_refresh_settings (
    user_id INTEGER PRIMARY KEY,
    interval_minutes INTEGER DEFAULT 60,
    last_update_ts TEXT,
    next_update_ts TEXT,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
