package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"recipe-web-server/internal/config"
	"recipe-web-server/internal/middleware"
	"recipe-web-server/internal/models"
	"recipe-web-server/internal/services"
	"recipe-web-server/internal/templates"
	"slices"
	"strconv"
	"strings"
	"time"
)

var errBadRequest = errors.New("bad request")

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
	recipeOverviews, err := h.recipeService.GetAllRecipeOverviews()
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("get recipe overviews: %w", err))
		return
	}

	type RecipeCardData struct {
		ID              int
		Title           string
		PrepTimeHours   int
		CookTimeMinutes int
		IsGreen         bool
		UpdatedAt       time.Time
	}

	recipeCardDatas := make([]RecipeCardData, len(recipeOverviews))

	for i, val := range recipeOverviews {
		recipeCardDatas[i] = RecipeCardData{
			ID:              val.ID,
			Title:           val.Title,
			PrepTimeHours:   val.PrepTimeSeconds / 3600,
			CookTimeMinutes: val.CookTimeSeconds / 60,
			IsGreen:         val.IsVegetarian(),
			UpdatedAt:       val.UpdatedAt,
		}
	}

	h.renderer.Render(w, r, "home", map[string]any{
		"Recipes": recipeCardDatas,
	})
}

func (h *Handler) ViewRecipes(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) ViewCreateRecipe(w http.ResponseWriter, r *http.Request) {
	h.renderer.Render(w, r, "recipe-form", map[string]any{})
}

func (h *Handler) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r)

	parsed, err := h.parseRecipeForm(r, user.ID)
	if err != nil {
		h.handleParseErr(w, r, err)
		return
	}

	// Image is required when creating a new recipe.
	if parsed.Image == nil && parsed.Errors["image"] == "" {
		parsed.Errors["image"] = "Bild krävs."
	}
	if parsed.Image != nil {
		defer parsed.Image.Close()
	}

	if len(parsed.Errors) > 0 {
		h.renderer.Render(w, r, "recipe-form", map[string]any{
			"Errors": parsed.Errors,
			"Recipe": parsed.Recipe,
		})
		return
	}

	id, err := h.recipeService.CreateRecipe(parsed.Recipe)
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("create recipe: %w", err))
		return
	}

	if err := services.ProcessAndSaveRecipeImage(id, parsed.Image, parsed.ImageExt); err != nil {
		h.recipeService.DeleteRecipeById(id)
		h.renderErrInternal(w, r, err)
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

	user, ok := middleware.GetUser(r)
	canManage := ok && (user.ID == recipe.OwnerID || user.Role == models.RoleAdmin)

	tags := make([]string, 0)
	for _, tag := range recipe.MealTypes {
		tags = append(tags, tag)
	}
	for _, tag := range recipe.DietaryTags {
		if slices.Contains(recipe.DietaryTags, models.TagVegan) && (tag == models.TagVegetarian || tag == models.TagMilkFree) {
			continue
		}
		tags = append(tags, tag)
	}
	for _, tag := range recipe.OtherTags {
		tags = append(tags, tag)
	}

	data := map[string]any{
		"Recipe":                       recipe,
		"RecipeTags":                   tags,
		"RecipePrepTimeHours":          recipe.PrepTimeSeconds / 3600,
		"RecipeCookTimeMinutes":        recipe.CookTimeSeconds / 60,
		"RecipeDescriptionParsed":      services.ParseMarkup(recipe.Description),
		"RecipeInstructionsParsed":     services.ParseMarkup(recipe.Instructions),
		"RecipePrepInstructionsParsed": services.ParseMarkup(recipe.PrepInstructions),
		"RecipeCreatedAtFormatted":     services.FormatDate(recipe.CreatedAt),
		"RecipeUpdatedAtFormatted":     services.FormatDate(recipe.UpdatedAt),
		"RecipeOwner":                  recipeOwner,
		"CanManage":                    canManage,
	}
	h.renderer.Render(w, r, "recipe", data)
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
		"IsEdit": true,
		"Recipe": recipe,
	}
	h.renderer.Render(w, r, "recipe-form", data)
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

func (h *Handler) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	user := middleware.MustGetUser(r)

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.renderErrBadRequest(w, r)
		return
	}

	existing, err := h.recipeService.GetRecipeById(id)
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("get recipe: %w", err))
		return
	}
	if existing == nil {
		h.renderErrPageNotFound(w, r)
		return
	}
	if existing.OwnerID != user.ID && user.Role != models.RoleAdmin {
		h.renderErrForbidden(w, r)
		return
	}

	parsed, err := h.parseRecipeForm(r, user.ID)
	if err != nil {
		h.handleParseErr(w, r, err)
		return
	}
	if parsed.Image != nil {
		defer parsed.Image.Close()
	}

	if len(parsed.Errors) > 0 {
		parsed.Recipe.ID = id
		h.renderer.Render(w, r, "recipe-form", map[string]any{
			"Errors": parsed.Errors,
			"Recipe": parsed.Recipe,
		})
		return
	}

	parsed.Recipe.ID = id
	if err := h.recipeService.UpdateRecipe(parsed.Recipe); err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("update recipe: %w", err))
		return
	}

	if parsed.Image != nil {
		if err := services.ProcessAndSaveRecipeImage(id, parsed.Image, parsed.ImageExt); err != nil {
			h.renderErrInternal(w, r, err)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", id), http.StatusSeeOther)
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
	h.renderer.Render(w, r, "login", data)
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
		Secure:   true, // TODO: set to true in production
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
	h.renderer.Render(w, r, "settings", data)
}

func (h *Handler) ViewAdminDashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "admin-dashboard", data)
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
	h.renderer.Render(w, r, "admin-users", data)
}

func (h *Handler) ViewCreateUser(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	h.renderer.Render(w, r, "admin-create-user", data)
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
	h.renderer.Render(w, r, "admin-manage-user", data)
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

// recipeFormResult holds everything parsed out of a recipe create/edit form.
// Image and ImageFileExt are only set when the user actually uploaded a file.
type recipeFormResult struct {
	Recipe   *models.Recipe
	Image    multipart.File
	ImageExt string
	Errors   map[string]string
}

// parseRecipeForm parses and validates the multipart form shared by both
// CreateRecipe and UpdateRecipe. ownerID is embedded in the returned Recipe.
// The caller is responsible for closing result.Image when it is non-nil.
func (h *Handler) parseRecipeForm(r *http.Request, ownerID int) (*recipeFormResult, error) {
	result := &recipeFormResult{
		Errors: map[string]string{},
	}

	if err := r.ParseMultipartForm(15 << 20); err != nil {
		result.Errors["form"] = "Ett problem uppstod. Om bilden är större än 10 MB, testa att ladda upp en mindre bild."
		return result, nil
	}

	imageFile, _, err := r.FormFile("image")
	if err == nil {
		buf := make([]byte, 512)
		if _, err := imageFile.Read(buf); err != nil && err != io.EOF {
			return nil, fmt.Errorf("read image header into buffer: %w", err)
		}
		switch contentType := http.DetectContentType(buf); contentType {
		case "image/jpeg":
			result.ImageExt = "jpg"
		case "image/png":
			result.ImageExt = "png"
		default:
			result.Errors["image"] = "Bara JPG, JPEG och PNG är tillåtna."
			imageFile.Close()
		}
		if result.ImageExt != "" {
			if _, err := imageFile.Seek(0, 0); err != nil {
				return nil, fmt.Errorf("seek image: %w", err)
			}
			result.Image = imageFile
		}
	}

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		result.Errors["title"] = "Receptnamn krävs."
	}

	servings := r.FormValue("servings")
	if strings.TrimSpace(servings) == "" {
		result.Errors["servings"] = "Antal portioner krävs."
	}

	instructions := r.FormValue("instructions")
	if strings.TrimSpace(instructions) == "" {
		result.Errors["instructions"] = "Instruktioner krävs."
	}

	description := r.FormValue("description")

	mealTypes := r.Form["meal-types[]"]
	diets := r.Form["diets[]"]
	tags := r.Form["tags[]"]

	if !services.ValidateRecipeTagList(mealTypes, models.AllowedMealTypes) ||
		!services.ValidateRecipeTagList(diets, models.AllowedDietaryTags) ||
		!services.ValidateRecipeTagList(tags, models.AllowedOtherTags) {
		return nil, fmt.Errorf("%w: tags", errBadRequest)
	}

	// --- Times ---
	cookTime, err := strconv.Atoi(r.FormValue("cook-time"))
	if err != nil {
		return nil, fmt.Errorf("%w: cook time", errBadRequest)
	}

	prepTimeHours, err := strconv.Atoi(r.FormValue("prep-time"))
	if err != nil {
		return nil, fmt.Errorf("%w: prep time", errBadRequest)
	}

	prepInstructions := ""
	if prepTimeHours > 0 {
		prepInstructions = r.FormValue("prep-instructions")
		if strings.TrimSpace(prepInstructions) == "" {
			result.Errors["prepInstructions"] = "Instruktion av förberedelsen krävs."
		}
	}

	// --- Ingredients ---
	var ingredientSections []models.IngredientSection
	if err := json.Unmarshal([]byte(r.FormValue("ingredients")), &ingredientSections); err != nil {
		return nil, fmt.Errorf("%w: ingredients", errBadRequest)
	}

	var ingredientErrs []string
	totalIngredients := 0
	hasEmptyHeading := false
	hasEmptyName := false

	for i, section := range ingredientSections {
		if i > 0 && strings.TrimSpace(section.Heading) == "" {
			hasEmptyHeading = true
		}
		for _, ingredient := range section.Ingredients {
			totalIngredients++
			if strings.TrimSpace(ingredient.Name) == "" {
				hasEmptyName = true
			}
		}
	}
	if totalIngredients == 0 {
		ingredientErrs = append(ingredientErrs, "Receptet måste ha minst en ingrediens.")
	}
	if hasEmptyHeading {
		ingredientErrs = append(ingredientErrs, "Alla sektioner utom den första måste ha en rubrik.")
	}
	if hasEmptyName {
		ingredientErrs = append(ingredientErrs, "Alla ingredienser måste ha ett namn.")
	}
	if len(ingredientErrs) > 0 {
		result.Errors["ingredients"] = strings.Join(ingredientErrs, " ")
	}

	result.Recipe = &models.Recipe{
		Title:              title,
		Description:        description,
		IngredientSections: ingredientSections,
		Instructions:       instructions,
		Servings:           servings,
		PrepTimeSeconds:    prepTimeHours * 3600,
		PrepInstructions:   prepInstructions,
		CookTimeSeconds:    cookTime,
		MealTypes:          mealTypes,
		DietaryTags:        diets,
		OtherTags:          tags,
		OwnerID:            ownerID,
	}

	return result, nil
}

// handleParseErr routes errors from parseRecipeForm to the right response.
// Bad-request errors (tampered form fields) get a 400; anything else is a 500.
func (h *Handler) handleParseErr(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, errBadRequest) {
		log.Printf("recipe form bad request: %v", err)
		h.renderErrBadRequest(w, r)
	} else {
		h.renderErrInternal(w, r, err)
	}
}
