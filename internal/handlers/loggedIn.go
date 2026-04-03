package handlers

import (
	"fmt"
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
		"Title": "Inställningar",
		"Path":  r.URL.Path,
		"User":  user,
	}
	h.renderer.Render(w, "settings", data)
}

func (h *LoggedInHandler) NewRecipePage(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)
	data := map[string]any{
		"Title": "Skapa nytt recept",
		"Path":  r.URL.Path,
		"User":  user,
	}
	h.renderer.Render(w, "recipe-new", data)
}

func (h *LoggedInHandler) NewRecipe(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	err := r.ParseMultipartForm(0)
	if err != nil {
		http.Error(w, "Kunde inte läsa formulär", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	mealTypes := r.Form["meal-types[]"]
	diets := r.Form["diets[]"]
	tags := r.Form["tags[]"]
	cookTime := r.FormValue("cook-time")
	prepTime := r.FormValue("prep-time")
	description := r.FormValue("description")
	image, _, err := r.FormFile("image")
	servings := r.FormValue("servings")
	ingredients := r.FormValue("ingredients")
	instructions := r.FormValue("instructions")
	if err != nil {
		http.Error(w, "Kunde inte läsa bilden", http.StatusBadRequest)
		return
	}

	fmt.Println("user:", user)
	fmt.Println("title:", title)
	fmt.Println("meal types:", mealTypes)
	fmt.Println("diets:", diets)
	fmt.Println("tags:", tags)
	fmt.Println("cook time:", cookTime)
	fmt.Println("prep time", prepTime)
	fmt.Println("desc", description)
	fmt.Println("image", image)
	fmt.Println("servings", servings)
	fmt.Println("ingredients", ingredients)
	fmt.Println("instructiosn", instructions)
}
