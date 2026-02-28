package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"recipe-web-server/internal/config"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
	"strconv"
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

func (h *AdminHandler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title": "Admin - Panel",
		"User":  user,
	}
	h.renderer.Render(w, "admin-dashboard", data)
}

func (h *AdminHandler) UsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunde inte läsa användare", http.StatusInternalServerError)
		return
	}
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title": "Admin - Hantera användare",
		"User":  user,
		"Users": users,
	}
	h.renderer.Render(w, "admin-users", data)
}

func (h *AdminHandler) CreateUserPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title": "Admin - Skapa användare",
		"User":  user,
	}
	h.renderer.Render(w, "admin-create-user", data)
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Redirect(w, r, "/admin/users/create?error=username_required", http.StatusSeeOther)
		return
	}

	displayName := r.FormValue("display-name")
	if displayName == "" {
		displayName = username
	}

	password := r.FormValue("password")
	if password == "" {
		http.Redirect(w, r, "/admin/users/create?error=password_required", http.StatusSeeOther)
		return
	}
	if len(password) < config.MinPasswordLength {
		http.Redirect(w, r, "/admin/users/create?error=password_too_short", http.StatusSeeOther)
		return
	}

	confirmPassword := r.FormValue("confirm-password")
	if password != confirmPassword {
		http.Redirect(w, r, "/admin/users/create?error=confirm_not_match", http.StatusSeeOther)
		return
	}

	role, ok := services.GetRole(r.FormValue("role"))
	if !ok {
		http.Redirect(w, r, "/admin/users/create?error=invalid_role", http.StatusSeeOther)
		return
	}

	_, err := h.userService.CreateUser(username, displayName, password, role)
	if err != nil {
		http.Error(w, "kunde inte skapa användare", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *AdminHandler) ManageUserPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	managedUser, err := h.userService.GetUser(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "användaren kunde inte hittas", http.StatusNotFound)
			return
		} else {
			http.Error(w, "internet serverfel", http.StatusInternalServerError)
			return
		}
	}
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title":       "Admin - Hantera användare",
		"User":        user,
		"ManagedUser": managedUser,
	}
	h.renderer.Render(w, "admin-manage-user", data)
}

func (h *AdminHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	newUsername := r.FormValue("username")
	err = h.userService.UpdateUsername(id, newUsername)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d", id), http.StatusSeeOther)
}

func (h *AdminHandler) UpdateDisplayName(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	displayName := r.FormValue("display-name")
	err = h.userService.UpdateDisplayName(id, displayName)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d", id), http.StatusSeeOther)
}

func (h *AdminHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm-password")
	err = h.userService.UpdatePassword(id, password, confirmPassword)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d", id), http.StatusSeeOther)
}

func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	role := r.FormValue("role")
	err = h.userService.UpdateRole(id, role)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/manage/%d", id), http.StatusSeeOther)
}
