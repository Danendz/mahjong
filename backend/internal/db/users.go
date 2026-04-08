package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

// User represents a guest user.
type User struct {
	ID           string
	Nickname     string
	SessionToken string
}

// CreateGuestUser creates a new guest user with a session token.
func (db *DB) CreateGuestUser(ctx context.Context, nickname string) (*User, error) {
	id := uuid.New().String()
	token, err := generateSessionToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	_, err = db.Pool.Exec(ctx,
		"INSERT INTO users (id, nickname, session_token) VALUES ($1, $2, $3)",
		id, nickname, token,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return &User{ID: id, Nickname: nickname, SessionToken: token}, nil
}

// GetUserByToken looks up a user by their session token.
func (db *DB) GetUserByToken(ctx context.Context, token string) (*User, error) {
	var user User
	err := db.Pool.QueryRow(ctx,
		"SELECT id, nickname, session_token FROM users WHERE session_token = $1",
		token,
	).Scan(&user.ID, &user.Nickname, &user.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// UpdateNickname updates a user's nickname.
func (db *DB) UpdateNickname(ctx context.Context, token, nickname string) error {
	_, err := db.Pool.Exec(ctx,
		"UPDATE users SET nickname = $1 WHERE session_token = $2",
		nickname, token,
	)
	return err
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
