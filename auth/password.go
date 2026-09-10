package auth

/*
|--------------------------------------------------------------------------
| Password
|--------------------------------------------------------------------------
|
| Implements Password as part of the auth package in the Gofreight
| framework. Key symbols: HashPassword, CheckPassword, User, Authenticate.
| 
| The auth package covers session login, password hashing, API token
| storage, OAuth callbacks, email verification, and password reset flows.
| 
| Controllers compose auth helpers with your User model; tokens and
| verification stores can be in-memory or database-backed.
| 
| Install scaffolding with gofreight make:auth and wire find-user
| callbacks in app/auth.
| 
| Symbols defined here include: DefaultCost (exported value); HashPassword
| (HashPassword hashes a plaintext password.); CheckPassword
| (CheckPassword compares a bcrypt hash with a plaintext password.);
| ErrInvalidCredentials (exported value); User (exported type);
| Authenticate (Authenticate checks email/password against a user
| record.).
| 
*/

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
