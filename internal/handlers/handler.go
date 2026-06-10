package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maltedickson/tomaternas/internal/apperrors"
	"github.com/maltedickson/tomaternas/internal/config"
	"github.com/maltedickson/tomaternas/internal/middleware"
	"github.com/maltedickson/tomaternas/internal/models"
	"github.com/maltedickson/tomaternas/internal/services"
	"github.com/maltedickson/tomaternas/internal/templates"
	"log"
	"net/http"
	"net/url"
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

func NewHandler(
	authService *services.AuthService,
	userService *services.UserService,
	recipeService *services.RecipeService,
	renderer *templates.Renderer,
) *Handler {
	return &Handler{
		userService:   userService,
		authService:   authService,
		recipeService: recipeService,
		renderer:      renderer,
	}
}

func (h *Handler) ViewHome(w http.ResponseWriter, r *http.Request) {
	recipeOverviews, err := h.recipeService.GetAllRecipeOverviews(r.Context())
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("getting recipe overviews: %w", err))
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
	if parsed != nil && parsed.Image != nil {
		defer parsed.Image.Close()
	}
	if err != nil {
		h.handleParseErr(w, r, err)
		return
	}
	id, err := h.recipeService.CreateRecipe(r.Context(), *parsed)
	if err != nil {
		var validationErr services.RecipeValidationError
		if errors.As(err, &validationErr) {
			h.renderer.Render(w, r, "recipe-form", map[string]any{
				"Errors": validationErr,
				"Recipe": &parsed.Recipe,
			})
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("creating recipe: %w", err))
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
	recipe, err := h.recipeService.GetRecipeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("getting recipe with ID %d: %w", id, err))
		return
	}

	recipeOwner, err := h.userService.GetUser(r.Context(), recipe.OwnerID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("getting user with id %d: %w", recipe.OwnerID, err))
		return
	}

	user, ok := middleware.GetUser(r)
	canManage := ok && services.CanManageRecipe(*user, *recipe)

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

	recipe, err := h.recipeService.GetRecipeByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("getting recipe with ID %d: %w", id, err))
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

	fromURL, err := url.QueryUnescape(r.FormValue("from"))
	if err != nil || !isInternalURL(fromURL) {
		h.renderErrBadRequest(w, r)
		return
	}
	if fromURL == "" {
		fromURL = "/"
	}

	user := middleware.MustGetUser(r)
	if err := h.recipeService.DeleteRecipeByID(r.Context(), id, *user); err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			h.renderErrForbidden(w, r)
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		h.renderErrInternal(w, r, err)
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

	parsed, err := h.parseRecipeForm(r, user.ID)
	if parsed != nil && parsed.Image != nil {
		defer parsed.Image.Close()
	}
	if err != nil {
		h.handleParseErr(w, r, err)
		return
	}

	if err := h.recipeService.UpdateRecipe(r.Context(), *user, id, *parsed); err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			h.renderErrForbidden(w, r)
			return
		}
		if errors.Is(err, apperrors.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		}
		var validationErr services.RecipeValidationError
		if errors.As(err, &validationErr) {
			h.renderer.Render(w, r, "recipe-form", map[string]any{
				"Errors": validationErr,
				"Recipe": &parsed.Recipe,
			})
			return
		}
		h.renderErrInternal(w, r, fmt.Errorf("updating recipe with id %d: %w", id, err))
		return
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

	session, err := h.authService.Login(r.Context(), username, password)
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
		h.authService.Logout(r.Context(), cookie.Value)
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
	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("getting all users: %w", err))
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

	_, err := h.userService.CreateUser(r.Context(), username, displayName, password, role)
	if err != nil {
		h.renderErrInternal(w, r, fmt.Errorf("creating user: %w", err))
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
	managedUser, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			h.renderErrPageNotFound(w, r)
			return
		} else {
			h.renderErrInternal(w, r, fmt.Errorf("getting user with ID %d: %w", id, err))
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
	err = h.userService.UpdateUsername(r.Context(), id, newUsername)
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
	err = h.userService.UpdateDisplayName(r.Context(), id, displayName)
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
	err = h.userService.UpdatePassword(r.Context(), id, password, confirmPassword)
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
	err = h.userService.UpdateRole(r.Context(), id, role)
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
	log.Printf("internal error: %s %s: %v", r.Method, r.URL.RequestURI(), err)
	h.renderer.RenderErr(w, r, http.StatusInternalServerError, "Något gick fel.")
}

func isInternalURL(urlStr string) bool {
	return urlStr == "" ||
		(strings.HasPrefix(urlStr, "/") && !strings.HasPrefix(urlStr, "//"))
}

// parseRecipeForm parses the multipart form used to create or update a recipe.
//
// The caller is responsible for closing .Image in the returned
// CreateRecipeInput.
// Even if an error is returned, the caller should check if an image was
// returned and in case it was, close it.
func (h *Handler) parseRecipeForm(r *http.Request, ownerID int) (*services.RecipeInput, error) {
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		return nil, fmt.Errorf("%w: parse multipart form", errBadRequest)
	}
	var result services.RecipeInput
	result.Recipe.OwnerID = ownerID
	imageFile, _, err := r.FormFile("image")
	if err == nil {
		result.Image = imageFile
	}
	result.Recipe.Title = r.FormValue("title")
	result.Recipe.Servings = r.FormValue("servings")
	result.Recipe.Instructions = r.FormValue("instructions")
	result.Recipe.Description = r.FormValue("description")
	result.Recipe.MealTypes = r.Form["meal-types[]"]
	result.Recipe.DietaryTags = r.Form["diets[]"]
	result.Recipe.OtherTags = r.Form["tags[]"]
	result.Recipe.CookTimeSeconds, err = strconv.Atoi(r.FormValue("cook-time"))
	if err != nil {
		return &result, fmt.Errorf("%w: cook time", errBadRequest)
	}
	prepTimeHours, err := strconv.Atoi(r.FormValue("prep-time"))
	if err != nil {
		return &result, fmt.Errorf("%w: prep time", errBadRequest)
	}
	result.Recipe.PrepTimeSeconds = prepTimeHours * 3600
	result.Recipe.PrepInstructions = r.FormValue("prep-instructions")
	if err := json.Unmarshal([]byte(r.FormValue("ingredients")), &result.Recipe.IngredientSections); err != nil {
		return &result, fmt.Errorf("%w: ingredients", errBadRequest)
	}
	return &result, nil
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
