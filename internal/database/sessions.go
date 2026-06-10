package database

import (
	"context"
	"fmt"

	"github.com/maltedickson/tomaternas/internal/models"
)

func (db *DB) CreateSession(ctx context.Context, s *models.Session) error {
	const query = `
		INSERT INTO sessions (token, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.ExecContext(ctx, query, s.Token, s.UserID, s.CreatedAt, s.ExpiresAt)
	if err != nil {
		return fmt.Errorf("db inserting session: %w", err)
	}
	return nil
}

func (db *DB) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	query := `
		SELECT token, user_id, created_at, expires_at
		FROM sessions
		WHERE token = ?
	`
	var s models.Session
	err := db.QueryRowContext(ctx, query, token).Scan(&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("db getting session: %w", err)
	}
	return &s, nil
}

func (db *DB) DeleteUserSessions(ctx context.Context, id int) error {
	const query = `
		DELETE FROM sessions
		WHERE user_id = ?
	`
	_, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db deleting session for user %d: %w", id, err)
	}
	return nil
}

func (db *DB) DeleteSession(ctx context.Context, token string) error {
	const query = `
		DELETE FROM sessions
		WHERE token = ?
	`
	_, err := db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("db deleting session: %w", err)
	}
	return nil
}
