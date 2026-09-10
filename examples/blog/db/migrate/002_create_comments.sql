-- --------------------------------------------------------------------------
-- 002 Create Comments
-- --------------------------------------------------------------------------
--
-- Database migration defining schema changes applied in order by gofreight
-- migrate.
--
-- Pair up/down migrations when altering columns; prefer blueprint Go
-- migrations for new projects.
--
-- --------------------------------------------------------------------------
-- 002 Create Comments
-- --------------------------------------------------------------------------
--
-- Database migration or schema SQL for versioned database changes.
--
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
