package middleware

import (
	"context"
	"net/http"
	"recipe-web-server/internal/models"
	"recipe-web-server/internal/services"
)

const SessionCookieName = "session_token"

func AuthMiddleware(authService *services.AuthService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				ctx := context.WithValue(r.Context(), IsAuthContextKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			user, err := authService.ValidateSession(cookie.Value)
			if err != nil {
				ctx := context.WithValue(r.Context(), IsAuthContextKey, false)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, IsAuthContextKey, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	return user, ok
}

func IsAuthenticated(r *http.Request) bool {
	isAuth, ok := r.Context().Value(IsAuthContextKey).(bool)
	return ok && isAuth
}
