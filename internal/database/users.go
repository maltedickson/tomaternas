package database

import (
	"recipe-web-server/internal/models"
)

func (db *DB) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (username, display_name, password_hash, role)
		VALUES (?, ?, ?, ?)
	`
	result, err := db.Exec(
		query,
		user.Username,
		user.DisplayName,
		user.PasswordHash,
		user.Role,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

func (db *DB) GetUserById(id int) (*models.User, error) {
	query := `
		SELECT id, username, display_name, password_hash, role, is_active, updated_at, created_at
		FROM users
		WHERE id = ?
	`
	var user models.User
	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.UpdatedAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, display_name, password_hash, role, is_active, updated_at, created_at
		FROM users
		WHERE username = ?
	`
	var user models.User
	err := db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.DisplayName,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.UpdatedAt,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *DB) GetAllUsers() ([]*models.User, error) {
	query := `
		SELECT id, username, display_name, password_hash, role, is_active, updated_at, created_at
		FROM users
	`

	rows, err := db.Query(query)
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
			&user.IsActive,
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

func (db *DB) SetUserActive(id int, active bool) error {
	var current bool
	err := db.QueryRow("SELECT is_active FROM users WHERE id = ?", id).Scan(&current)
	if err != nil {
		return err
	}
	if current == active {
		return nil
	}
	query := `
		UPDATE users
		SET is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err = db.Exec(query, active, id)
	return err
}
