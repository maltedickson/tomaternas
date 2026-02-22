package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(username, password string) (*models.Session, error) {
	user, err := s.db.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	const sessionTimeToLive = 6 * time.Hour

	session := &models.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(sessionTimeToLive),
		CreatedAt: time.Now(),
	}

	if err := s.db.CreateSession(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) Logout(token string) error {
	return s.db.DeleteSession(token)
}

func (s *AuthService) ValidateSession(token string) (*models.User, error) {
	session, err := s.db.GetSessionByToken(token)
	if err != nil {
		return nil, err
	}
	if session.IsExpired() {
		s.db.DeleteSession(token)
		return nil, errors.New("session expired")
	}
	user, err := s.db.GetUserById(session.UserID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
