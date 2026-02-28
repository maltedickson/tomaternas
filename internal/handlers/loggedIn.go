package handlers

import (
	"net/http"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
)

type LoggedInHandler struct {
	userService *services.UserService
	renderer    *templates.Renderer
}

func NewLoggedInHandler(userService *services.UserService, renderer *templates.Renderer) *LoggedInHandler {
	return &LoggedInHandler{
		userService: userService,
		renderer:    renderer,
	}
}

func (h *LoggedInHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title": "Home",
		"Path":  r.URL.Path,
		"User":  user,
	}
	h.renderer.Render(w, "settings", data)
}
