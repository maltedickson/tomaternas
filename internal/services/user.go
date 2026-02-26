package services

import (
	"errors"
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

	p := &params{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 1,
		saltLength:  16,
		keyLength:   32,
	}

	passwordHash := generateFromPassword(password, p)

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
	return s.db.GetUserById(id)
}

func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.db.GetAllUsers()
}
