package services

import (
	"database/sql"
	"errors"
	"fmt"
	"recipe-web-server/internal/config"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
)

type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

var validRoles = map[string]models.UserRole{
	string(models.RoleAdmin): models.RoleAdmin,
	string(models.RoleUser):  models.RoleUser,
}

func GetRole(roleStr string) (models.UserRole, bool) {
	role, ok := validRoles[roleStr]
	return role, ok
}

func (s *UserService) CreateUser(username, displayName, password string, role models.UserRole) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password required")
	}

	_, err := s.db.GetUserByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	passwordHash := generateFromPassword(password, defaultParams)

	user := &models.User{
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
		Role:         role,
	}

	if err := s.db.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(id int) (*models.User, error) {
	user, err := s.db.GetUserById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return user, nil
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.db.GetAllUsers()
}

func (s *UserService) UpdateUsername(id int, newUsername string) error {
	_, err := s.db.GetUserByUsername(newUsername)
	if err == nil {
		return errors.New("username already exists")
	}
	return s.db.UpdateUsername(id, newUsername)
}

func (s *UserService) UpdateDisplayName(id int, displayName string) error {
	return s.db.UpdateDisplayName(id, displayName)
}

func (s *UserService) UpdatePassword(id int, password, confirmPassword string) error {
	if password != confirmPassword {
		return errors.New("confirm_not_match")
	}
	if len(password) < config.MinPasswordLength {
		return errors.New("password_too_short")
	}
	hash := generateFromPassword(password, defaultParams)
	return s.db.UpdatePasswordHash(id, hash)
}

func (s *UserService) UpdateRole(id int, role string) error {
	return s.db.UpdateRole(id, role)
}
