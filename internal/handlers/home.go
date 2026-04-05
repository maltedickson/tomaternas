package handlers

import (
	"net/http"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
	"strconv"
)

type HomeHandler struct {
	userService   *services.UserService
	recipeService *services.RecipeService
	renderer      *templates.Renderer
}

func NewHomeHandler(userService *services.UserService, recipeService *services.RecipeService, renderer *templates.Renderer) *HomeHandler {
	return &HomeHandler{
		userService:   userService,
		recipeService: recipeService,
		renderer:      renderer,
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

func (h *HomeHandler) ViewRecipePage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ogiltigt id", http.StatusBadRequest)
		return
	}
	recipe, err := h.recipeService.GetRecipeById(id)
	if err != nil {
		http.Error(w, "Receptet kunde inte hittas", http.StatusNotFound)
		return
	}

	recipeOwner, err := h.userService.GetUser(recipe.OwnerID)
	if err != nil {
		http.Error(w, "internt serverfel", http.StatusInternalServerError)
		return
	}

	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title":       recipe.Title,
		"Path":        r.URL.Path,
		"User":        user,
		"Recipe":      recipe,
		"RecipeOwner": recipeOwner,
	}
	h.renderer.Render(w, "recipe", data)
}
