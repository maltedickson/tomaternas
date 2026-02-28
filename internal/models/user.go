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
	PasswordHash string
	Role         UserRole
	UpdatedAt    time.Time
	CreatedAt    time.Time
}
