package services

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/disintegration/imaging"
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

func ProcessAndSaveRecipeImage(recipeId int, uploadedFile multipart.File, imageExt string) error {
	uploadDir := filepath.Join("data", "uploads", "recipes")

	pathForOriginalImage := filepath.Join(uploadDir, fmt.Sprintf("%d_original.%s", recipeId, imageExt))

	err := saveFileToDisk(uploadedFile, pathForOriginalImage)
	if err != nil {
		return fmt.Errorf("failed to save original image to path \"%s\": %w", pathForOriginalImage, err)
	}

	srcImage, err := imaging.Decode(uploadedFile)
	if err != nil {
		return fmt.Errorf("failed to decode uploaded image: %w", err)
	}

	sizeLarge := 600
	sizeThumb := 350

	imageLarge2x := imaging.Fill(srcImage, sizeLarge*2, sizeLarge*2, imaging.Center, imaging.Lanczos)

	imageThumb := imaging.Fill(srcImage, sizeThumb, sizeThumb, imaging.Center, imaging.Lanczos)
	imageThumb2x := imaging.Fill(srcImage, sizeThumb*2, sizeThumb*2, imaging.Center, imaging.Lanczos)

	pathLarge2x := filepath.Join(uploadDir, fmt.Sprintf("%d_lg@2x.jpg", recipeId))
	pathThumb1x := filepath.Join(uploadDir, fmt.Sprintf("%d_thumb@1x.jpg", recipeId))
	pathThumb2x := filepath.Join(uploadDir, fmt.Sprintf("%d_thumb@2x.jpg", recipeId))

	err = imaging.Save(imageLarge2x, pathLarge2x, imaging.JPEGQuality(80))
	if err != nil {
		return fmt.Errorf("failed to save jpeg lg@2x: %w", err)
	}

	err = imaging.Save(imageThumb, pathThumb1x, imaging.JPEGQuality(60))
	if err != nil {
		return fmt.Errorf("failed to save jpeg thumb@1x: %w", err)
	}

	err = imaging.Save(imageThumb2x, pathThumb2x, imaging.JPEGQuality(60))
	if err != nil {
		return fmt.Errorf("failed to save jpeg thumb@2x: %w", err)
	}

	return nil
}

func saveFileToDisk(file multipart.File, filepath string) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek image: %w", err)
	}

	dst, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		return err
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek image: %w", err)
	}

	return nil
}
