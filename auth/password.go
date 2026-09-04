package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = bcrypt.DefaultCost

// HashPassword hashes a plaintext password.
func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	return string(b), err
}

// CheckPassword compares a bcrypt hash with a plaintext password.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

var ErrInvalidCredentials = errors.New("invalid credentials")

// User represents a minimal authenticatable user.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
}

// Authenticate checks email/password against a user record.
func Authenticate(u *User, password string) error {
	if u == nil || !CheckPassword(u.PasswordHash, password) {
		return ErrInvalidCredentials
	}
	return nil
}
