package services

import (
	"errors"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(username, displayName, password string, isAdmin bool) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password required")
	}

	_, err := s.db.GetUserByUsername(username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:       username,
		DisplayName:    displayName,
		HashedPassword: hashedPassword,
		IsAdmin:        isAdmin,
		CreatedAt:      time.Now(),
	}

	if err := s.db.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUser(id int) (*models.User, error) {
	return s.db.GetUserById(id)
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.db.GetAllUsers()
}

func (s *UserService) DeleteUser(id int) error {
	if err := s.db.DeleteUserSessions(id); err != nil {
		return err
	}
	return s.db.DeleteUser(id)
}
