package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"recipe-web-server/internal/config"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/models"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	userService   *services.UserService
	authService   *services.AuthService
	recipeService *services.RecipeService
	renderer      *templates.Renderer
}

func NewHandler(authService *services.AuthService, userService *services.UserService, recipeService *services.RecipeService, renderer *templates.Renderer) *Handler {
	return &Handler{
		userService:   userService,
		authService:   authService,
		recipeService: recipeService,
		renderer:      renderer,
	}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Metoden är ej tillåten", http.StatusMethodNotAllowed)
		return
	}

	recipeOverviews, err := h.recipeService.GetAllRecipeOverviews()
	if err != nil {
		http.Error(w, "Kunde inte hämta recept", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"RecipeOverviews": recipeOverviews,
	}
	h.renderer.Render(w, r, "home", "Hem", data)
}

func (h *Handler) ViewRecipePage(w http.ResponseWriter, r *http.Request) {
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

	dataDirectory := "data"
	imageMatches, err := filepath.Glob(filepath.Join(dataDirectory, "uploads", "recipes", fmt.Sprintf("%d.*", id)))
	if err != nil || len(imageMatches) == 0 {
		http.Error(w, "kunde inte läsa bilden", http.StatusInternalServerError)
		return
	}
	imagePath := imageMatches[0]
	imageSrc, err := filepath.Rel(dataDirectory, imagePath)
	if err != nil {
		http.Error(w, "kunde inte läsa bilden", http.StatusInternalServerError)
		return
	}

	prepTimeFormatted := ""
	if recipe.PrepTimeSeconds > 0 {
		prepTimeFormatted = fmt.Sprintf("%d h", recipe.PrepTimeSeconds/3600)
	}

	cookTimeFormatted := fmt.Sprintf("%d min", recipe.CookTimeSeconds/60)

	data := map[string]any{
		"Recipe":                   recipe,
		"RecipePrepTimeFormatted":  prepTimeFormatted,
		"RecipeCookTimeFormatted":  cookTimeFormatted,
		"RecipeDescriptionParsed":  services.ParseMarkup(recipe.Description),
		"RecipeInstructionsParsed": services.ParseMarkup(recipe.Instructions),
		"RecipeCreatedAtFormatted": services.FormatDate(recipe.CreatedAt),
		"RecipeUpdatedAtFormatted": services.FormatDate(recipe.UpdatedAt),
		"RecipeImageSrc":           imageSrc,
		"RecipeOwner":              recipeOwner,
	}
	h.renderer.Render(w, r, "recipe", recipe.Title, data)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
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
		"Error":      errMsg,
		"ReturnPath": returnPath,
	}
	h.renderer.Render(w, r, "login", "Logga in", data)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "settings", "Inställningar", data)
}

func (h *Handler) NewRecipePage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "recipe-new", "Skapa nytt recept", data)
}

func (h *Handler) NewRecipe(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) DashboardPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "admin-dashboard", "Admin - Panel", data)
}

func (h *Handler) UsersPage(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		http.Error(w, "Kunde inte läsa användare", http.StatusInternalServerError)
		return
	}
	data := map[string]any{
		"Users": users,
	}
	h.renderer.Render(w, r, "admin-users", "Admin - Hantera användare", data)
}

func (h *Handler) CreateUserPage(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "admin-create-user", "Admin - Skapa användare", data)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) ManageUserPage(w http.ResponseWriter, r *http.Request) {
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
	data := map[string]any{
		"ManagedUser": managedUser,
	}
	h.renderer.Render(w, r, "admin-manage-user", "Admin - Hantera användare", data)
}

func (h *Handler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateDisplayName(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
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

func isInternalURL(urlStr string) bool {
	return urlStr == "" ||
		(strings.HasPrefix(urlStr, "/") && !strings.HasPrefix(urlStr, "//"))
}
