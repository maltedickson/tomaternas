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
	data := map[string]any{
		"Title":           "Home",
		"IsAuthenticated": middleware.IsAuthenticated(r),
		"Path":            r.URL.Path,
	}
	h.renderer.Render(w, "home", data)
}

func (h *HomeHandler) OtherPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":           "Annan sida",
		"IsAuthenticated": middleware.IsAuthenticated(r),
		"Path":            r.URL.Path,
	}
	h.renderer.Render(w, "other-page", data)
}
