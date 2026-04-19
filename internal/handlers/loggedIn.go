package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/models"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
	"strconv"
)

type LoggedInHandler struct {
	userService   *services.UserService
	recipeService *services.RecipeService
	renderer      *templates.Renderer
}

func NewLoggedInHandler(userService *services.UserService, recipeService *services.RecipeService, renderer *templates.Renderer) *LoggedInHandler {
	return &LoggedInHandler{
		userService:   userService,
		recipeService: recipeService,
		renderer:      renderer,
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

	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		http.Error(w, "Kunde inte läsa formulär", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	mealTypes := r.Form["meal-types[]"]
	diets := r.Form["diets[]"]
	tags := r.Form["tags[]"]
	description := r.FormValue("description")
	servings := r.FormValue("servings")
	instructions := r.FormValue("instructions")

	cookTimeString := r.FormValue("cook-time")
	cookTime, err := strconv.Atoi(cookTimeString)
	if err != nil {
		http.Error(w, "Kunde inte läsa tid", http.StatusBadRequest)
		return
	}

	prepTimeString := r.FormValue("prep-time")
	prepTime, err := strconv.Atoi(prepTimeString)
	if err != nil {
		http.Error(w, "Kunde inte läsa förberedelsetid", http.StatusBadRequest)
		return
	}

	ingredientsString := r.FormValue("ingredients")
	var ingredientSections []models.IngredientSection
	err = json.Unmarshal([]byte(ingredientsString), &ingredientSections)
	if err != nil {
		http.Error(w, "Kunde inte läsa ingredienser", http.StatusBadRequest)
		return
	}

	imageFile, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Bild krävs", http.StatusBadRequest)
		return
	}
	defer imageFile.Close()

	imageBuff := make([]byte, 512)
	if _, err := imageFile.Read(imageBuff); err != nil {
		http.Error(w, "Kunde inte läsa bilden", http.StatusInternalServerError)
		return
	}
	imageFileType := http.DetectContentType(imageBuff)

	var imageFileExt string
	switch imageFileType {
	case "image/jpeg":
		imageFileExt = "jpg"
	case "image/png":
		imageFileExt = "png"
	default:
		http.Error(w, "Bara JPG och PNG är tillåtna", http.StatusUnsupportedMediaType)
		return
	}

	imageFile.Seek(0, 0)

	recipe := &models.Recipe{
		Title:              title,
		Description:        description,
		IngredientSections: ingredientSections,
		Instructions:       instructions,
		Servings:           servings,
		PrepTimeSeconds:    prepTime,
		CookTimeSeconds:    cookTime,
		MealTypes:          mealTypes,
		DietaryTags:        diets,
		OtherTags:          tags,
		OwnerID:            user.ID,
	}

	id, err := h.recipeService.CreateRecipe(recipe)

	if err != nil {
		http.Error(w, "Kunde inte skapa receptet"+err.Error(), http.StatusInternalServerError)
		return
	}

	imageFileName := fmt.Sprintf("%d.%s", id, imageFileExt)
	imageFilePath := filepath.Join("data", "uploads", "recipes", imageFileName)

	imageDestinationFile, err := os.Create(imageFilePath)
	if err != nil {
		http.Error(w, "Kunde inte spara filen", http.StatusInternalServerError)
		return
	}
	defer imageDestinationFile.Close()

	if _, err := io.Copy(imageDestinationFile, imageFile); err != nil {
		h.recipeService.DeleteRecipeById(id)
		os.Remove(imageFilePath)

		http.Error(w, "Kunde inte spara bilden", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/r/%d", id), http.StatusSeeOther)
}
