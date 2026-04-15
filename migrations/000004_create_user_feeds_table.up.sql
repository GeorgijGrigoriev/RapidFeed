CREATE TABLE IF NOT EXISTS user_feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    feed_url TEXT,
    title TEXT,
    category TEXT,
    FOREIGN KEY(user_id) REFERENCES users(id)
);
