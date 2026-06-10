package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/maltedickson/tomaternas/internal/apperrors"
	"github.com/maltedickson/tomaternas/internal/config"
	"github.com/maltedickson/tomaternas/internal/database"
	"github.com/maltedickson/tomaternas/internal/models"
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

func (s *UserService) CreateUser(ctx context.Context, username, displayName, password string, role models.UserRole) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password required")
	}

	_, err := s.db.GetUserByUsername(ctx, username)
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

	if err := s.db.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id int) (*models.User, error) {
	user, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.db.GetAllUsers(ctx)
}

func (s *UserService) UpdateUsername(ctx context.Context, id int, newUsername string) error {
	_, err := s.db.GetUserByUsername(ctx, newUsername)
	if err == nil {
		return errors.New("username already exists")
	}
	return s.db.UpdateUsername(ctx, id, newUsername)
}

func (s *UserService) UpdateDisplayName(ctx context.Context, id int, displayName string) error {
	return s.db.UpdateDisplayName(ctx, id, displayName)
}

func (s *UserService) UpdatePassword(ctx context.Context, id int, password, confirmPassword string) error {
	if password != confirmPassword {
		return errors.New("confirm_not_match")
	}
	if len(password) < config.MinPasswordLength {
		return errors.New("password_too_short")
	}
	hash := generateFromPassword(password, defaultParams)
	return s.db.UpdatePasswordHash(ctx, id, hash)
}

func (s *UserService) UpdateRole(ctx context.Context, id int, role string) error {
	return s.db.UpdateRole(ctx, id, role)
}
