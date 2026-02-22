package models

import (
	"time"
)

type User struct {
	ID             int
	Username       string
	DisplayName    string
	HashedPassword []byte
	IsAdmin        bool
	CreatedAt      time.Time
}
