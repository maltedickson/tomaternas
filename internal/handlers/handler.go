package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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

func (h *Handler) ViewHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		h.renderErrMethodNotAllowed(w, r)
		return
	}

	recipeOverviews, err := h.recipeService.GetAllRecipeOverviews()
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("get recipe overviews: %w", err))
		return
	}

	data := map[string]any{
		"RecipeOverviews": recipeOverviews,
	}
	h.renderer.Render(w, r, "home", "Hem", data)
}

func (h *Handler) ViewRecipes(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) ViewCreateRecipe(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "recipe-new", "Skapa nytt recept", data)
}

func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.GetUser(r)

	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		h.renderErrBadRequest(w, r)
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
		h.renderErrBadRequest(w, r)
		return
	}

	prepTimeString := r.FormValue("prep-time")
	prepTimeHours, err := strconv.Atoi(prepTimeString)
	if err != nil {
		h.renderErrBadRequest(w, r)
		return
	}
	prepTime := prepTimeHours * 3600

	ingredientsString := r.FormValue("ingredients")
	var ingredientSections []models.IngredientSection
	err = json.Unmarshal([]byte(ingredientsString), &ingredientSections)
	if err != nil {
		h.renderErrBadRequest(w, r)
		return
	}

	imageFile, _, err := r.FormFile("image")
	if err != nil {
		h.renderErrBadRequest(w, r)
		return
	}
	defer imageFile.Close()

	imageBuff := make([]byte, 512)
	if _, err := imageFile.Read(imageBuff); err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("read image into buffer: %w", err))
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
		// TODO: give nice error message and allow user to fix their form data
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
		h.renderErrInternal(w, r, fmt.Errorf("create recipe: %w", err))
		return
	}

	imageFileName := fmt.Sprintf("%d.%s", id, imageFileExt)
	imageFilePath := filepath.Join("data", "uploads", "recipes", imageFileName)

	imageDestinationFile, err := os.Create(imageFilePath)
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("create image file: %w", err))
		return
	}
	defer imageDestinationFile.Close()

	if _, err := io.Copy(imageDestinationFile, imageFile); err != nil {
		h.recipeService.DeleteRecipeById(id)
		os.Remove(imageFilePath)

		h.renderErrInternal(w, r, fmt.Errorf("copy image to created file: %w", err))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", id), http.StatusSeeOther)
}

func (h *Handler) ViewRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}
	recipe, err := h.recipeService.GetRecipeById(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("get recipe by id (%d): %w", id, err))
		return
	}

	recipeOwner, err := h.userService.GetUser(recipe.OwnerID)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("get user by id (%d): %w", recipe.OwnerID, err))
		return
	}

	dataDirectory := "data"
	imageMatches, err := filepath.Glob(filepath.Join(dataDirectory, "uploads", "recipes", fmt.Sprintf("%d.*", id)))
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("get images for recipe with id (%d): %w", id, err))
		return
	}
	if len(imageMatches) == 0 {
		h.renderErrInternal(w, r, fmt.Errorf("get images for recipe with id (%d): no images found", id))
		return
	}
	imagePath := imageMatches[0]
	imageSrc, err := filepath.Rel(dataDirectory, imagePath)
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("calculate image src: %w", err))
		return
	}

	prepTimeFormatted := ""
	if recipe.PrepTimeSeconds > 0 {
		prepTimeFormatted = fmt.Sprintf("%d h", recipe.PrepTimeSeconds/3600)
	}

	cookTimeFormatted := fmt.Sprintf("%d min", recipe.CookTimeSeconds/60)

	user, ok := middleware.GetUser(r)
	canManage := ok && (user.ID == recipe.OwnerID || user.Role == models.RoleAdmin)

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
		"CanManage":                canManage,
	}
	h.renderer.Render(w, r, "recipe", recipe.Title, data)
}

func (h *Handler) ViewEditRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}

	recipe, err := h.recipeService.GetRecipeById(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("get recipe by id (%d): %w", id, err))
		return
	}

	data := map[string]any{
		"Recipe": recipe,
	}
	h.renderer.Render(w, r, "recipe-new", "Redigera recept", data)
}

func (h *Handler) DeleteRecipe(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}

	recipe, err := h.recipeService.GetRecipeById(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("get recipe by id (%d): %w", id, err))
		return
	}

	user := middleware.MustGetUser(r)

	havePermission := user.ID == recipe.OwnerID || user.Role == models.RoleAdmin
	if !havePermission {
		h.renderErrForbidden(w, r)
		return
	}

	fromURL, err := url.QueryUnescape(r.FormValue("from"))
	if err != nil {
		h.renderErrBadRequest(w, r)
		return
	}

	if err := h.recipeService.DeleteRecipeById(id); err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("delete recipe by id (%d): %w", id, err))
		return
	}

	if fromURL == "" {
		fromURL = "/"
	}

	if !isInternalURL(fromURL) {
		fromURL = "/"
	}

	http.Redirect(w, r, fromURL, http.StatusSeeOther)
}

func (h *Handler) ViewLogin(w http.ResponseWriter, r *http.Request) {
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
		h.renderErrBadRequest(w, r)
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

func (h *Handler) ViewSettings(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "settings", "Inställningar", data)
}

func (h *Handler) ViewAdminDashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "admin-dashboard", "Admin - Panel", data)
}

func (h *Handler) ViewUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("get all users: %w", err))
		return
	}
	data := map[string]any{
		"Users": users,
	}
	h.renderer.Render(w, r, "admin-users", "Admin - Hantera användare", data)
}

func (h *Handler) ViewCreateUser(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "admin-create-user", "Admin - Skapa användare", data)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		http.Redirect(w, r, "/admin/users/new?error=username_required", http.StatusSeeOther)
		return
	}

	displayName := r.FormValue("display-name")
	if displayName == "" {
		displayName = username
	}

	password := r.FormValue("password")
	if password == "" {
		http.Redirect(w, r, "/admin/users/new?error=password_required", http.StatusSeeOther)
		return
	}
	if len(password) < config.MinPasswordLength {
		http.Redirect(w, r, "/admin/users/new?error=password_too_short", http.StatusSeeOther)
		return
	}

	confirmPassword := r.FormValue("confirm-password")
	if password != confirmPassword {
		http.Redirect(w, r, "/admin/users/new?error=confirm_not_match", http.StatusSeeOther)
		return
	}

	role, ok := services.GetRole(r.FormValue("role"))
	if !ok {
		http.Redirect(w, r, "/admin/users/new?error=invalid_role", http.StatusSeeOther)
		return
	}

	_, err := h.userService.CreateUser(username, displayName, password, role)
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("create user: %w", err))
		return
	}

	http.Redirect(w, r, "/admin/users", http.StatusSeeOther)
}

func (h *Handler) ViewUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}
	managedUser, err := h.userService.GetUser(id)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		} else {
			h.renderErrInternal(w, r, fmt.Errorf("get user by id (%d): %w", id, err))
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
		h.renderErrPageNotFound(w, r)
		return
	}
	newUsername := r.FormValue("username")
	err = h.userService.UpdateUsername(id, newUsername)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit", id), http.StatusSeeOther)
}

func (h *Handler) UpdateDisplayName(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}
	displayName := r.FormValue("display-name")
	err = h.userService.UpdateDisplayName(id, displayName)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit", id), http.StatusSeeOther)
}

func (h *Handler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm-password")
	err = h.userService.UpdatePassword(id, password, confirmPassword)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit", id), http.StatusSeeOther)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.renderErrPageNotFound(w, r)
		return
	}
	role := r.FormValue("role")
	err = h.userService.UpdateRole(id, role)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit?error=%s", id, url.QueryEscape(err.Error())), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/users/%d/edit", id), http.StatusSeeOther)
}

func (h *Handler) renderErrMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderErr(w, r, http.StatusMethodNotAllowed, "Metoden är inte tillåten.")
}

func (h *Handler) renderErrPageNotFound(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderErr(w, r, http.StatusNotFound, "Sidan kunde inte hittas.")
}

func (h *Handler) renderErrBadRequest(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderErr(w, r, http.StatusBadRequest, "Dålig förfrågan.")
}

func (h *Handler) renderErrForbidden(w http.ResponseWriter, r *http.Request) {
	h.renderer.RenderErr(w, r, http.StatusForbidden, "Du saknar behörighet.")
}

func (h *Handler) renderErrInternal(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal error: %v", err)
	h.renderer.RenderErr(w, r, http.StatusInternalServerError, "Något gick fel.")
}

func isInternalURL(urlStr string) bool {
	return urlStr == "" ||
		(strings.HasPrefix(urlStr, "/") && !strings.HasPrefix(urlStr, "//"))
}
