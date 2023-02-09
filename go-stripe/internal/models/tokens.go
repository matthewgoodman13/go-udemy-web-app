package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const (
	ScopeAuthentication = "authentication"
)

// Token represents an authentication token.
type Token struct {
	PlainText string    `json:"token"`
	UserID    int64     `json:"-"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// GenerateToken generates a new token for the given user ID, expiry time and scope.
func GenerateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

// InsertToken inserts a token into the database.
func (m *DBModel) InsertToken(token *Token, u User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Delete any existing tokens for the user
	query := `DELETE FROM tokens WHERE user_id = ?`
	_, err := m.DB.ExecContext(ctx, query, token.UserID)
	if err != nil {
		return err
	}

	// Insert new token
	query = `INSERT INTO tokens (user_id, name, email, token_hash, expiry, created_at, updated_at) values (?, ?, ?, ?, ?, ?, ?)`
	_, err = m.DB.ExecContext(ctx, query, token.UserID, u.LastName, u.Email, token.Hash, token.Expiry, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

// GetUserForToken returns the user for the given token string.
func (m *DBModel) GetUserForToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))
	var user User

	query := `SELECT
				u.id, u.first_name, u.last_name, u.email
			  FROM
			    users u
			  INNER JOIN tokens t ON t.user_id = u.id
			  WHERE
			    t.token_hash = ? AND t.expiry > ?
	`
	err := m.DB.QueryRowContext(ctx, query, tokenHash[:], time.Now()).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
