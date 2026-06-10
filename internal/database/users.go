package database

import (
	"context"
	"fmt"

	"github.com/maltedickson/tomaternas/internal/models"
)

func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (username, display_name, password_hash, role)
		VALUES (?, ?, ?, ?)
	`
	result, err := db.ExecContext(
		ctx,
		query,
		user.Username,
		user.DisplayName,
		user.PasswordHash,
		user.Role,
	)
	if err != nil {
		return fmt.Errorf("db insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("db get last insert ID for inserted user: %w", err)
	}

	user.ID = int(id)
	return nil
}

func (db *DB) GetUserById(ctx context.Context, id int) (*models.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE id = ?
	`
	var user models.User
	err := db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Role,
		&user.UpdatedAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT *
		FROM users
		WHERE username = ?
	`
	var user models.User
	err := db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Role,
		&user.UpdatedAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	query := `
		SELECT *
		FROM users
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.DisplayName,
			&user.PasswordHash,
			&user.Role,
			&user.UpdatedAt,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

func (db *DB) UpdateUsername(ctx context.Context, id int, username string) error {
	query := `
		UPDATE users
		SET username = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.ExecContext(ctx, query, username, id)
	return err
}

func (db *DB) UpdateDisplayName(ctx context.Context, id int, displayName string) error {
	var current string
	err := db.QueryRowContext(ctx, "SELECT display_name FROM users WHERE id = ?", id).Scan(&current)
	if err != nil {
		return err
	}
	if current == displayName {
		return nil
	}
	query := `
		UPDATE users
		SET display_name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = db.ExecContext(ctx, query, displayName, id)
	return err
}

func (db *DB) UpdatePasswordHash(ctx context.Context, id int, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.ExecContext(ctx, query, passwordHash, id)
	return err
}

func (db *DB) UpdateRole(ctx context.Context, id int, role string) error {
	var current string
	err := db.QueryRowContext(ctx, "SELECT role FROM users WHERE id = ?", id).Scan(&current)
	if err != nil {
		return err
	}
	if current == role {
		return nil
	}
	query := `
		UPDATE users
		SET role = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = db.ExecContext(ctx, query, role, id)
	return err
}
