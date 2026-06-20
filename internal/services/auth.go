package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/maltedickson/tomaternas/internal/database"
	"github.com/maltedickson/tomaternas/internal/models"
	"github.com/maltedickson/tomaternas/internal/passwords"
)

type AuthService struct {
	db *database.DB
}

func NewAuthService(db *database.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*models.Session, error) {
	user, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	ok, err := passwords.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
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

	if err := s.db.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.db.DeleteSession(ctx, token)
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	session, err := s.db.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if session.IsExpired() {
		s.db.DeleteSession(ctx, token)
		return nil, errors.New("session expired")
	}
	user, err := s.db.GetUserByID(ctx, session.UserID)
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
