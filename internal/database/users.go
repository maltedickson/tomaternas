package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/maltedickson/tomaternas/internal/apperrors"
	"github.com/maltedickson/tomaternas/internal/models"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
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
		return fmt.Errorf("db inserting user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("db getting ID for inserted user: %w", err)
	}

	user.ID = int(id)
	return nil
}

func (db *DB) GetUserByID(ctx context.Context, id int) (*models.User, error) {
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("db getting user %d: %w", id, err)
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("db getting user with username %s: %w", username, err)
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
		return nil, fmt.Errorf("db selecting users: %w", err)
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
			return nil, fmt.Errorf("db scanning user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("db iterating users: %w", err)
	}

	return users, nil
}

func (db *DB) UpdateUsername(ctx context.Context, id int, username string) error {
	query := `
		UPDATE users
		SET username = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.ExecContext(ctx, query, username, id)
	if err != nil {
		return fmt.Errorf("db updating username for user %d: %w", id, err)
	}
	return nil
}

// UpdateDisplayName returns ErrConflict if the display name is already taken.
func (db *DB) UpdateDisplayName(ctx context.Context, id int, displayName string) error {
	query := `
		UPDATE users
		SET display_name = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	if _, err := db.ExecContext(ctx, query, displayName, id); err != nil {
		var sqliteErr *sqlite.Error
		if errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("db updating display name for user %d: %w", id, err)
	}
	return nil
}

func (db *DB) UpdatePasswordHash(ctx context.Context, id int, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.ExecContext(ctx, query, passwordHash, id)
	if err != nil {
		return fmt.Errorf("db updating password hash for user %d: %w", id, err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("db updating role for user %d: %w", id, err)
	}
	return nil
}
