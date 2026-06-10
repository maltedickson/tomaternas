package services

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"recipe-web-server/internal/apperrors"
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

type RecipeInput struct {
	Image  multipart.File
	Recipe models.Recipe
}

type RecipeValidationError struct {
	ImageMissing                 bool
	ImageContentTypeNotJpegOrPng bool

	TitleEmpty bool

	MealTypesHaveInvalidValues bool
	MealTypesHaveDuplicates    bool
	MealTypesEmpty             bool

	DietaryTagsHaveInvalidValues            bool
	DietaryTagsHaveDuplicates               bool
	DietaryTagsContainVeganButNotVegetarian bool
	DietaryTagsContainVeganButNotMilkFree   bool

	OtherTagsHaveInvalidValues bool
	OtherTagsHaveDuplicates    bool

	PrepTimeNegative                         bool
	PrepTimePositiveAndPrepInstructionsEmpty bool
	CookTimeNotPositive                      bool

	ServingsEmpty bool

	NonFirstIngredientSectionHeadingEmpty bool
	IngredientNameEmpty                   bool
	IngredientSectionEmpty                bool
	IngredientSectionsEmpty               bool

	InstructionsEmpty bool
}

func (r RecipeValidationError) Error() string {
	return "recipe validation failed"
}

func (s *RecipeService) CreateRecipe(ctx context.Context, input RecipeInput) (int, error) {
	normalizeRecipeInput(&input.Recipe)
	imageFileExt, err := validateRecipeInput(input, true)
	if err != nil {
		return 0, err
	}
	id, err := s.db.CreateRecipe(ctx, &input.Recipe)
	if err != nil {
		return 0, fmt.Errorf("database error: %w", err)
	}
	if err := ProcessAndSaveRecipeImage(id, input.Image, imageFileExt); err != nil {
		s.DeleteRecipeById(ctx, id)
		return 0, fmt.Errorf("saving image: %w", err)
	}
	return id, nil
}

// UpdateRecipe updates a recipe. Only the owner or an admin may update a
// recipe.
// Returns ErrNotFound, ErrForbidden, or a RecipeValidationError on failure.
func (s *RecipeService) UpdateRecipe(ctx context.Context, user models.User, recipeID int, input RecipeInput) error {
	normalizeRecipeInput(&input.Recipe)
	existingRecipe, err := s.db.GetRecipeById(ctx, recipeID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return err
		}
		return fmt.Errorf("getting recipe with id %d from database: %w", recipeID, err)
	}
	if existingRecipe.OwnerID != user.ID && user.Role != models.RoleAdmin {
		return apperrors.ErrForbidden
	}
	imageFileExt, err := validateRecipeInput(input, false)
	if err != nil {
		return err
	}
	if input.Image != nil {
		if err := ProcessAndSaveRecipeImage(recipeID, input.Image, imageFileExt); err != nil {
			return fmt.Errorf("saving image: %w", err)
		}
	}
	input.Recipe.ID = recipeID
	if err = s.db.UpdateRecipe(ctx, &input.Recipe); err != nil {
		return fmt.Errorf("updating recipe with id %d: %w", recipeID, err)
	}
	return nil
}

// GetRecipeById returns the recipe with the specified ID. If no such recipe exists, GetRecipeById returns [ErrNotFound]. If some error occurs, GetRecipeById returns an error.
func (s *RecipeService) GetRecipeById(ctx context.Context, id int) (*models.Recipe, error) {
	recipe, err := s.db.GetRecipeById(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return recipe, nil
}

func (s *RecipeService) GetAllRecipeOverviews(ctx context.Context) ([]models.RecipeOverview, error) {
	return s.db.GetAllRecipeOverviews(ctx)
}

func (s *RecipeService) DeleteRecipeById(ctx context.Context, id int) error {
	return s.db.DeleteRecipeById(ctx, id)
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

func normalizeRecipeInput(recipeInput *models.Recipe) {
	recipeInput.Title = strings.TrimSpace(recipeInput.Title)
	recipeInput.Description = strings.TrimSpace(recipeInput.Description)
	for i := range recipeInput.IngredientSections {
		section := &recipeInput.IngredientSections[i]
		section.Heading = strings.TrimSpace(section.Heading)
		for j := range section.Ingredients {
			ingredient := &section.Ingredients[j]
			ingredient.Name = strings.TrimSpace(ingredient.Name)
			ingredient.Amount = strings.TrimSpace(ingredient.Amount)
		}
	}
	recipeInput.Instructions = strings.TrimSpace(recipeInput.Instructions)
	recipeInput.Servings = strings.TrimSpace(recipeInput.Servings)
	recipeInput.PrepInstructions = strings.TrimSpace(recipeInput.PrepInstructions)
	trimStringSlice(recipeInput.MealTypes)
	trimStringSlice(recipeInput.DietaryTags)
	trimStringSlice(recipeInput.OtherTags)
}

func trimStringSlice(s []string) {
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
}

// validateRecipeInput validates the provided recipe input and returns the
// image file extension if an image is present. If any fields are invalid, a
// RecipeValidationError is returned describing all validation failures found.
//
// requireImage controls whether a missing image is treated as a validation
// error, which should be true for create and false for update.
//
// If an image is provided, it is opened to validate its content type.
// The caller is responsible for closing input.Image if it is non-nil,
// regardless of whether a validation error is returned.
func validateRecipeInput(input RecipeInput, requireImage bool) (imageFileExt string, err error) {
	var v RecipeValidationError

	// ----- Image -----

	if input.Image != nil {
		buf := make([]byte, 512)
		if _, err := input.Image.Read(buf); err != nil && err != io.EOF {
			return "", fmt.Errorf("reading image header into buffer: %w", err)
		}
		switch ct := http.DetectContentType(buf); ct {
		case "image/jpeg":
			imageFileExt = "jpg"
		case "image/png":
			imageFileExt = "png"
		default:
			v.ImageContentTypeNotJpegOrPng = true
		}
		if _, err := input.Image.Seek(0, io.SeekStart); err != nil {
			return "", fmt.Errorf("seeking image: %w", err)
		}
	} else if requireImage {
		v.ImageMissing = true
	}

	// ----- Title -----

	if input.Recipe.Title == "" {
		v.TitleEmpty = true
	}

	// ----- Tags -----

	v.MealTypesHaveInvalidValues, v.MealTypesHaveDuplicates = ValidateRecipeTagList(input.Recipe.MealTypes, models.AllowedMealTypes)
	v.MealTypesEmpty = len(input.Recipe.MealTypes) == 0

	v.DietaryTagsHaveInvalidValues, v.DietaryTagsHaveDuplicates = ValidateRecipeTagList(input.Recipe.DietaryTags, models.AllowedDietaryTags)
	if slices.Contains(input.Recipe.DietaryTags, models.TagVegan) {
		if !slices.Contains(input.Recipe.DietaryTags, models.TagVegetarian) {
			v.DietaryTagsContainVeganButNotVegetarian = true
		}
		if !slices.Contains(input.Recipe.DietaryTags, models.TagMilkFree) {
			v.DietaryTagsContainVeganButNotMilkFree = true
		}
	}

	v.OtherTagsHaveInvalidValues, v.OtherTagsHaveDuplicates = ValidateRecipeTagList(input.Recipe.OtherTags, models.AllowedOtherTags)

	// ----- Time -----

	if input.Recipe.PrepTimeSeconds < 0 {
		v.PrepTimeNegative = true
	} else if input.Recipe.PrepTimeSeconds > 0 && input.Recipe.PrepInstructions == "" {
		v.PrepTimePositiveAndPrepInstructionsEmpty = true
	}
	if input.Recipe.CookTimeSeconds <= 0 {
		v.CookTimeNotPositive = true
	}

	// ----- Servings -----

	if input.Recipe.Servings == "" {
		v.ServingsEmpty = true
	}

	// ----- Ingredients -----

	if len(input.Recipe.IngredientSections) == 0 {
		v.IngredientSectionsEmpty = true
	} else {
		for i, section := range input.Recipe.IngredientSections {
			if i > 0 && section.Heading == "" {
				v.NonFirstIngredientSectionHeadingEmpty = true
			}
			ingredientCount := 0
			for _, ingredient := range section.Ingredients {
				ingredientCount++
				if ingredient.Name == "" {
					v.IngredientNameEmpty = true
				}
			}
			if ingredientCount == 0 {
				v.IngredientSectionEmpty = true
			}
		}
	}

	// ----- Instructions -----

	if input.Recipe.Instructions == "" {
		v.InstructionsEmpty = true
	}

	// ----- Done validating, returning -----

	if v != (RecipeValidationError{}) {
		return imageFileExt, v
	}

	return imageFileExt, nil
}

func ValidateRecipeTagList(tags []string, allowedTags []string) (hasInvalidValues bool, containsDuplicates bool) {
	seen := make(map[string]struct{})
	for _, tag := range tags {
		if _, ok := seen[tag]; ok {
			containsDuplicates = true
		}
		if !slices.Contains(allowedTags, tag) {
			hasInvalidValues = true
		}
	}
	return hasInvalidValues, containsDuplicates
}

func ProcessAndSaveRecipeImage(recipeId int, uploadedFile multipart.File, imageExt string) error {
	uploadDir := filepath.Join("data", "uploads", "recipes")

	pathForOriginalImage := filepath.Join(uploadDir, fmt.Sprintf("%d_original.%s", recipeId, imageExt))

	err := saveFileToDisk(uploadedFile, pathForOriginalImage)
	if err != nil {
		return fmt.Errorf("saving original image to path \"%s\": %w", pathForOriginalImage, err)
	}

	srcImage, err := imaging.Decode(uploadedFile, imaging.AutoOrientation(true))
	if err != nil {
		return fmt.Errorf("decoding uploaded image: %w", err)
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
		return fmt.Errorf("saving jpeg lg@2x: %w", err)
	}

	err = imaging.Save(imageThumb, pathThumb1x, imaging.JPEGQuality(60))
	if err != nil {
		return fmt.Errorf("saving jpeg thumb@1x: %w", err)
	}

	err = imaging.Save(imageThumb2x, pathThumb2x, imaging.JPEGQuality(60))
	if err != nil {
		return fmt.Errorf("saving jpeg thumb@2x: %w", err)
	}

	return nil
}

func saveFileToDisk(file multipart.File, filepath string) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking image: %w", err)
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
		return fmt.Errorf("seeking image: %w", err)
	}

	return nil
}
