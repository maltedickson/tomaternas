package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/maltedickson/tomaternas/internal/apperrors"
	"github.com/maltedickson/tomaternas/internal/database"
	"github.com/maltedickson/tomaternas/internal/models"
)

type ReviewService struct {
	db *database.DB
}

func NewReviewService(db *database.DB) *ReviewService {
	return &ReviewService{db: db}
}

func (s *ReviewService) GetReviewsForRecipe(ctx context.Context, recipeID int) ([]models.Review, error) {
	return s.db.GetReviewsForRecipe(ctx, recipeID)
}

type ReviewValidationError struct {
	RatingNotBetween1And5 bool
}

func (vErr ReviewValidationError) Error() string {
	return "review validation failed"
}

// CreateReview creates a review for a recipe.
// Returns ErrNotFound if recipe does not exist,
// ErrAlreadyExists if user already reviewed the recipe,
// and ReviewValidationError on validation failure.
func (s *ReviewService) CreateReview(ctx context.Context, rating int, comment string, recipeID int, userID int) (int, error) {
	_, err := s.db.GetRecipeByID(ctx, recipeID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return 0, err
		}
		return 0, fmt.Errorf("getting recipe to verify that it exists: %w", err)
	}
	existingReviews, err := s.db.GetReviewsForRecipe(ctx, recipeID)
	if err != nil {

	}
	if slices.ContainsFunc(existingReviews, func(existingReview models.Review) bool {
		return existingReview.OwnerID == userID
	}) {
		return 0, apperrors.ErrAlreadyExists
	}

	comment = strings.TrimSpace(comment)

	var validationErr ReviewValidationError
	if rating < 1 || rating > 5 {
		validationErr.RatingNotBetween1And5 = true
	}
	if validationErr != (ReviewValidationError{}) {
		return 0, validationErr
	}

	reviewID, err := s.db.CreateReview(ctx, &models.Review{
		RecipeID: recipeID,
		OwnerID:  userID,
		Rating:   rating,
		Comment:  comment,
	})
	if err != nil {
		return 0, err
	}
	return reviewID, nil
}
