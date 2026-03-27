CREATE TABLE IF NOT EXISTS feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    link TEXT,
    date TIMESTAMP,
    source TEXT,
    description TEXT,
    feed_url TEXT
);
