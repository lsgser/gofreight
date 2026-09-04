-- Migration: create_comments
CREATE TABLE IF NOT EXISTS comments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	post_id INTEGER NOT NULL,
	body TEXT NOT NULL,
	author TEXT,
	created_at TEXT DEFAULT (datetime('now')),
	updated_at TEXT DEFAULT (datetime('now')),
	FOREIGN KEY (post_id) REFERENCES posts(id)
);
