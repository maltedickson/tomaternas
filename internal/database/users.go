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

func (db *DB) UpdateUsername(id int, username string) error {
	query := `
		UPDATE users
		SET username = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, username, id)
	return err
}

func (db *DB) UpdateDisplayName(id int, displayName string) error {
	var current string
	err := db.QueryRow("SELECT display_name FROM users WHERE id = ?", id).Scan(&current)
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
	_, err = db.Exec(query, displayName, id)
	return err
}

func (db *DB) UpdatePasswordHash(id int, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := db.Exec(query, passwordHash, id)
	return err
}

func (db *DB) UpdateRole(id int, role string) error {
	var current string
	err := db.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&current)
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
	_, err = db.Exec(query, role, id)
	return err
}
