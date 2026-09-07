package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/lsgser/gofreight/database"
)

// DatabaseTokenStore persists API tokens in the database.
type DatabaseTokenStore struct{}

func NewDatabaseTokenStore() *DatabaseTokenStore {
	return &DatabaseTokenStore{}
}

func (s *DatabaseTokenStore) Create(userID int64, name string, expiresAt time.Time) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	_, err := database.DB().ExecContext(context.Background(),
		`INSERT INTO api_tokens (user_id, name, token, expires_at) VALUES (?, ?, ?, ?)`,
		userID, name, token, expiresAt.Format(time.RFC3339),
	)
	return token, err
}

func (s *DatabaseTokenStore) Validate(token string) (int64, bool) {
	var userID int64
	var expires sql.NullString
	err := database.DB().QueryRowContext(context.Background(),
		`SELECT user_id, expires_at FROM api_tokens WHERE token = ?`, token,
	).Scan(&userID, &expires)
	if err != nil {
		return 0, false
	}
	if expires.Valid && expires.String != "" {
		t, err := time.Parse(time.RFC3339, expires.String)
		if err == nil && time.Now().After(t) {
			return 0, false
		}
	}
	return userID, true
}

func (s *DatabaseTokenStore) Revoke(token string) error {
	_, err := database.DB().ExecContext(context.Background(),
		`DELETE FROM api_tokens WHERE token = ?`, token,
	)
	return err
}
