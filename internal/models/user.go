package models

import (
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID           int
	Username     string
	DisplayName  string
	PasswordHash []byte
	Role         UserRole
	IsActive     bool
	UpdatedAt    time.Time
	CreatedAt    time.Time
}
