package handlers

import (
	"net/http"
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

func (h *AdminHandler) DeleteUserPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	user, err := h.userService.GetUser(id)
	if err != nil {
		http.Error(w, "användaren kunde inte hittas", http.StatusNotFound)
		return
	}
	data := map[string]any{
		"Title":           "Admin - Ta bort användare",
		"IsAuthenticated": middleware.IsAuthenticated(r),
		"User":            *user,
	}
	h.renderer.Render(w, "admin-delete-user", data)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	err = h.userService.DeleteUser(id)
	if err != nil {
		http.Error(w, "kunde inte ta bort användaren", http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}
