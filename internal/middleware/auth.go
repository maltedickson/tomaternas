package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"github.com/maltedickson/tomaternas/internal/models"
	"github.com/maltedickson/tomaternas/internal/services"
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

			user, err := authService.ValidateSession(r.Context(), cookie.Value)
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

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuth, ok := r.Context().Value(IsAuthContextKey).(bool)
		if !ok || !isAuth {
			http.Redirect(w, r, "/login?return="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*models.User)
		if !ok || user.Role != models.RoleAdmin {
			http.Error(w, "Förbjuden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func GetUser(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	return user, ok
}

func MustGetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	if !ok {
		panic(errors.New("expected client to be logged in, but found no user"))
	}
	return user
}

func IsAuthenticated(r *http.Request) bool {
	isAuth, ok := r.Context().Value(IsAuthContextKey).(bool)
	return ok && isAuth
}

func IsAdmin(r *http.Request) bool {
	user, ok := GetUser(r)
	return ok && user.Role == models.RoleAdmin
}
