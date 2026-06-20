package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/maltedickson/tomaternas/internal/apperrors"
	"github.com/maltedickson/tomaternas/internal/config"
	"github.com/maltedickson/tomaternas/internal/database"
	"github.com/maltedickson/tomaternas/internal/models"
	"github.com/maltedickson/tomaternas/internal/passwords"
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

	passwordHash := passwords.Hash(password, passwords.DefaultParams)

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

type DisplayNameValidationErr struct {
	DisplayNameNotBetween1And64Chars bool
}

func (vErr DisplayNameValidationErr) Error() string {
	return "display name validation failed"
}

// UpdateDisplayName returns ErrConflict if the display name is already taken,
// or a DisplayNameValidationErr on validation failure.
func (s *UserService) UpdateDisplayName(
	ctx context.Context,
	id int,
	displayName string,
) error {
	displayName = strings.TrimSpace(displayName)
	var validationErr DisplayNameValidationErr
	if len(displayName) < 1 || len(displayName) > 32 {
		validationErr.DisplayNameNotBetween1And64Chars = true
	}
	if validationErr != (DisplayNameValidationErr{}) {
		return validationErr
	}
	if err := s.db.UpdateDisplayName(ctx, id, displayName); err != nil {
		return err
	}
	return nil
}

type PasswordValidationErr struct {
	PasswordTooShort bool
}

func (vErr PasswordValidationErr) Error() string {
	return "password validation failed"
}

// UpdatePassword returns ErrInvalidCredentials if the supplied current password
// is incorrect, and a PasswordValidationErr on validation failure.
func (s *UserService) UpdatePassword(
	ctx context.Context,
	id int,
	currentPassword,
	newPassword string,
) error {
	if err := validateNewPassword(newPassword); err != nil {
		var vErr PasswordValidationErr
		if errors.As(err, &vErr) {
			return err
		}
		return fmt.Errorf("validating new password: %w", err)
	}

	user, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		return err
	}
	isCurrentPasswordCorrect, err := passwords.ComparePasswordAndHash(
		currentPassword,
		user.PasswordHash,
	)
	if err != nil {
		return err
	}
	if !isCurrentPasswordCorrect {
		return apperrors.ErrInvalidCredentials
	}

	newHash := passwords.Hash(newPassword, passwords.DefaultParams)
	if err := s.db.UpdatePasswordHash(ctx, id, newHash); err != nil {
		return err
	}
	return nil
}

func (s *UserService) AdminUpdatePassword(ctx context.Context, id int, password, confirmPassword string) error {
	if password != confirmPassword {
		return errors.New("confirm_not_match")
	}
	if len(password) < config.MinPasswordLength {
		return errors.New("password_too_short")
	}
	hash := passwords.Hash(password, passwords.DefaultParams)
	return s.db.UpdatePasswordHash(ctx, id, hash)
}

func (s *UserService) UpdateRole(ctx context.Context, id int, role string) error {
	return s.db.UpdateRole(ctx, id, role)
}

func validateNewPassword(password string) error {
	var vErr PasswordValidationErr
	if len(password) < config.MinPasswordLength {
		vErr.PasswordTooShort = true
	}
	if vErr != (PasswordValidationErr{}) {
		return vErr
	}
	return nil
}
