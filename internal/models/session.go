package models

import (
	"time"
)

type Session struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (s *Session) IsExpired() bool {
	return time.Now().Compare(s.ExpiresAt) == 1
}
