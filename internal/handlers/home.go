package handlers

import (
	"net/http"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/templates"
)

type HomeHandler struct {
	renderer *templates.Renderer
}

func NewHomeHandler(renderer *templates.Renderer) *HomeHandler {
	return &HomeHandler{
		renderer: renderer,
	}
}

func (h *HomeHandler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Metoden är ej tillåten", http.StatusMethodNotAllowed)
		return
	}
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title": "Home",
		"Path":  r.URL.Path,
		"User":  user,
	}
	h.renderer.Render(w, "home", data)
}
