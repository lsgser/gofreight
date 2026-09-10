package generator

/*
|--------------------------------------------------------------------------
| Scaffold
|--------------------------------------------------------------------------
|
| Implements Scaffold as part of the generator package in the Gofreight
| framework. Key symbols: Auth.
| 
| The generator package powers gofreight new and all make:* scaffolds.
| 
| It writes idiomatic directory layouts, GFT views, migrations, tests, and
| auth stubs from templates.
| 
| CLI handlers in cmd/gofreight call into this package; templates live
| primarily in templates.go.
| 
| Symbols defined here include: Auth (Auth generates user model,
| migration, and login routes snippet.).
| 
*/

import (
	"os"
	"path/filepath"
)

// Auth generates user model, migration, and login routes snippet.
func Auth(appPath string) error {
	fields := map[string]string{
		"email":    "email",
		"password": "string",
		"role":     "string",
	}
	if err := Model(appPath, "User", fields); err != nil {
		return err
	}

	dir := filepath.Join(appPath, "db", "migrate")
	os.MkdirAll(dir, 0755)

	migrations := map[string]string{
		"002_create_password_reset_tokens.sql": `CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL,
  token TEXT NOT NULL,
  expires_at TEXT NOT NULL,
  created_at TEXT DEFAULT (datetime('now'))
);`,
		"002_create_password_reset_tokens_down.sql": `DROP TABLE IF EXISTS password_reset_tokens;`,
		"003_create_api_tokens.sql": `CREATE TABLE IF NOT EXISTS api_tokens (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  token TEXT NOT NULL UNIQUE,
  expires_at TEXT,
  created_at TEXT DEFAULT (datetime('now'))
);`,
		"003_create_api_tokens_down.sql": `DROP TABLE IF EXISTS api_tokens;`,
		"004_add_email_verified_to_users.sql": `ALTER TABLE users ADD COLUMN email_verified_at TEXT;`,
		"004_add_email_verified_to_users_down.sql": `ALTER TABLE users DROP COLUMN email_verified_at;`,
	}
	for name, sql := range migrations {
		_ = os.WriteFile(filepath.Join(dir, name), []byte(sql), 0644)
	}

	seedPath := filepath.Join(appPath, "db", "seeds", "users.sql")
	os.MkdirAll(filepath.Dir(seedPath), 0755)
	seed := `-- Seed admin user (password: secret123 — change in production)
INSERT INTO users (email, password, role) VALUES ('admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin');
`
	if err := os.WriteFile(seedPath, []byte(seed), 0644); err != nil {
		return err
	}

	snippetPath := filepath.Join(appPath, "config", "auth_routes.txt")
	snippet := `Auth scaffolding installed — login/register routes wired in routes/auth.go`
	if err := os.WriteFile(snippetPath, []byte(snippet), 0644); err != nil {
		return err
	}
	return installAuthStarter(appPath)
}
