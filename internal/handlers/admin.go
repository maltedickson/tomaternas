package handlers

import (
	"net/http"
	"recipe-web-server/internal/config"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
)

type AdminHandler struct {
	userService *services.UserService
	renderer    *templates.Renderer
}

func NewAdminHandler(userService *services.UserService, renderer *templates.Renderer) *AdminHandler {
	return &AdminHandler{
		userService: userService,
		renderer:    renderer,
	}
}

func (h *AdminHandler) UsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunde inte läsa användare", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Title":           "Admin - Hantera användare",
		"IsAuthenticated": middleware.IsAuthenticated(r),
		"Users":           users,
	}
	h.renderer.Render(w, "admin-users", data)
}

func (h *AdminHandler) CreateUserPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":           "Admin - Skapa användare",
		"IsAuthenticated": middleware.IsAuthenticated(r),
	}
	h.renderer.Render(w, "admin-create-user", data)
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Redirect(w, r, "/admin/users/create?error=username_required", http.StatusBadRequest)
		return
	}

	displayName := r.FormValue("display-name")
	if displayName == "" {
		displayName = username
	}

	password := r.FormValue("password")
	if password == "" {
		http.Redirect(w, r, "/admin/users/create?error=password_required", http.StatusBadRequest)
		return
	}
	if len(password) < config.MinPasswordLength {
		http.Redirect(w, r, "/admin/users/create?error=password_too_short", http.StatusBadRequest)
		return
	}

	confirmPassword := r.FormValue("confirm-password")
	if password != confirmPassword {
		http.Redirect(w, r, "/admin/users/create?error=confirm_not_match", http.StatusBadRequest)
		return
	}

	role, ok := services.GetRole(r.FormValue("role"))
	if !ok {
		http.Redirect(w, r, "/admin/users/create?error=invalid_role", http.StatusBadRequest)
		return
	}

	_, err := h.userService.CreateUser(username, displayName, password, role)
	if err != nil {
		http.Error(w, "kunde inte skapa användare", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
