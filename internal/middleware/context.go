package middleware

type contextKey string

const (
	UserContextKey   contextKey = "user"
	IsAuthContextKey contextKey = "isAuthenticated"
)
