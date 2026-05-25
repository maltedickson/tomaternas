package services

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
	"regexp"
	"slices"
	"strings"
	"time"
)

type RecipeService struct {
	db *database.DB
}

func NewRecipeService(db *database.DB) *RecipeService {
	return &RecipeService{db: db}
}

func (s *RecipeService) CreateRecipe(recipe *models.Recipe) (int, error) {
	return s.db.CreateRecipe(recipe)
}

func (s *RecipeService) UpdateRecipe(recipe *models.Recipe) error {
	return s.db.UpdateRecipe(recipe)
}

// GetRecipeById returns the recipe with the specified ID. If no such recipe exists, GetRecipeById returns [ErrNotFound]. If some error occurs, GetRecipeById returns an error.
func (s *RecipeService) GetRecipeById(id int) (*models.Recipe, error) {
	recipe, err := s.db.GetRecipeById(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return recipe, nil
}

func (s *RecipeService) GetAllRecipeOverviews() ([]models.RecipeOverview, error) {
	return s.db.GetAllRecipeOverviews()
}

func (s *RecipeService) DeleteRecipeById(id int) error {
	return s.db.DeleteRecipeById(id)
}

func ParseMarkup(markup string) template.HTML {
	if markup == "" {
		return ""
	}

	markup = strings.ReplaceAll(markup, "\r\n", "\n")
	markup = strings.ReplaceAll(markup, "\r", "\n")

	paragraphs := regexp.MustCompile("\n\n+").Split(markup, -1)

	var result strings.Builder
	for _, p := range paragraphs {
		if p == "" {
			continue
		}
		result.WriteString("<p>")
		lines := strings.Split(p, "\n")
		for j, line := range lines {
			if j > 0 {
				result.WriteString("<br>")
			}
			result.WriteString(template.HTMLEscapeString(line))
		}
		result.WriteString("</p>")
	}
	return template.HTML(result.String())
}

func FormatDate(date time.Time) string {
	return date.Format("2006-01-02")
}

func ValidateRecipeTagList(tags []string, allowedTags []string) bool {
	seen := make(map[string]struct{})
	for _, tag := range tags {
		if _, ok := seen[tag]; ok {
			return false
		}
		if !slices.Contains(allowedTags, tag) {
			return false
		}
	}
	return true
}
