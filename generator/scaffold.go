package generator

import (
	"os"
	"path/filepath"
)

// Auth generates user model, migration, and login routes snippet.
func Auth(appPath string) error {
	fields := map[string]string{
		"email":    "string",
		"password": "string",
		"role":     "string",
	}
	if err := Model(appPath, "User", fields); err != nil {
		return err
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
	snippet := `// Auth routes — add to config/routes.go:
// r.Post("/login", controller.Handler(auth.Login(auth.DefaultLoginConfig(findUserByEmail))))
// r.Post("/logout", controller.Handler(auth.Logout("current_user_id", "/")))
`
	return os.WriteFile(snippetPath, []byte(snippet), 0644)
}
