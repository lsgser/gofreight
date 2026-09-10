-- --------------------------------------------------------------------------
-- 001 Create Posts
-- --------------------------------------------------------------------------
--
-- Database migration defining schema changes applied in order by gofreight
-- migrate.
--
-- Pair up/down migrations when altering columns; prefer blueprint Go
-- migrations for new projects.
--
-- --------------------------------------------------------------------------
-- 001 Create Posts
-- --------------------------------------------------------------------------
--
-- Database migration or schema SQL for versioned database changes.
--
-- Migration: create_posts
CREATE TABLE IF NOT EXISTS posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title VARCHAR(255) NOT NULL,
	body TEXT NOT NULL,
	created_at TEXT DEFAULT (datetime('now')),
	updated_at TEXT DEFAULT (datetime('now'))
);
