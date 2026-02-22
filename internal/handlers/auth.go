package handlers

import (
	"net/http"
	"net/url"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
	"strings"
	"time"
)

type AuthHandler struct {
	authService *services.AuthService
	renderer    *templates.Renderer
}

func NewAuthHandler(authService *services.AuthService, renderer *templates.Renderer) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		renderer:    renderer,
	}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if middleware.IsAuthenticated(r) {
		returnURL := r.URL.Query().Get("return")
		if returnURL == "" || !isInternalURL(returnURL) {
			returnURL = "/"
		}
		http.Redirect(w, r, returnURL, http.StatusSeeOther)
		return
	}
	errMsg := ""
	switch r.URL.Query().Get("error") {
	case "invalid_credentials":
		errMsg = "Fel användarnamn eller lösenord."
	}
	returnPath := r.URL.Query().Get("return")
	data := map[string]any{
		"Title":      "Logga in",
		"Error":      errMsg,
		"ReturnPath": returnPath,
	}
	h.renderer.Render(w, "login", data)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	returnURL, err := url.QueryUnescape(r.FormValue("return"))
	if err != nil {
		http.Error(w, "invalid return paramter", http.StatusBadRequest)
		return
	}

	session, err := h.authService.Login(username, password)
	if err != nil {
		redirectURL := "/login?error=invalid_credentials"
		if returnURL != "" {
			redirectURL += "&return=" + url.QueryEscape(returnURL)
		}
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    session.Token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // TODO: set to true in production
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	if returnURL == "" {
		returnURL = "/"
	}

	if !isInternalURL(returnURL) {
		returnURL = "/"
	}

	http.Redirect(w, r, returnURL, http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	returnPath := r.FormValue("return")

	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err == nil {
		h.authService.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		MaxAge:   -1,
		Path:     "/",
	})

	if returnPath == "" {
		returnPath = "/"
	}

	if !isInternalURL(returnPath) {
		returnPath = "/"
	}

	http.Redirect(w, r, returnPath, http.StatusSeeOther)
}

func isInternalURL(urlStr string) bool {
	return urlStr == "" ||
		(strings.HasPrefix(urlStr, "/") && !strings.HasPrefix(urlStr, "//"))
}
