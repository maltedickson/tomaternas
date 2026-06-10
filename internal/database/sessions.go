package database

import (
	"context"
	"github.com/maltedickson/tomaternas/internal/models"
)

func (db *DB) CreateSession(ctx context.Context, s *models.Session) error {
	const query = `
		INSERT INTO sessions (token, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.ExecContext(ctx, query, s.Token, s.UserID, s.CreatedAt, s.ExpiresAt)
	if err != nil {
		return err
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
		return nil, err
	}
	return &s, nil
}

func (db *DB) DeleteUserSessions(ctx context.Context, id int) error {
	const query = `
		DELETE FROM sessions
		WHERE user_id = ?
	`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func (db *DB) DeleteSession(ctx context.Context, token string) error {
	const query = `
		DELETE FROM sessions
		WHERE token = ?
	`
	_, err := db.ExecContext(ctx, query, token)
	return err
}
