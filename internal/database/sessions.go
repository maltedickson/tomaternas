package database

import (
	"recipe-web-server/internal/models"
)

func (db *DB) CreateSession(s *models.Session) error {
	const query = `
		INSERT INTO sessions (token, user_id, created_at, expires_at)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(query, s.Token, s.UserID, s.CreatedAt, s.ExpiresAt)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) GetSessionByToken(token string) (*models.Session, error) {
	query := `
		SELECT token, user_id, created_at, expires_at
		FROM sessions
		WHERE token = ?
	`
	var s models.Session
	err := db.QueryRow(query, token).Scan(&s.Token, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (db *DB) DeleteUserSessions(id int) error {
	const query = `
		DELETE FROM sessions
		WHERE user_id = ?
	`
	_, err := db.Exec(query, id)
	return err
}

func (db *DB) DeleteSession(token string) error {
	const query = `
		DELETE FROM sessions
		WHERE token = ?
	`
	_, err := db.Exec(query, token)
	return err
}
