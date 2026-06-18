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

func (s *ReviewService) GetRecipeReviewItems(ctx context.Context, recipeID int) ([]models.RecipeReviewItem, error) {
	return s.db.GetRecipeReviewItems(ctx, recipeID)
}

type ReviewValidationError struct {
	RatingNotBetween1And5  bool
	CommentLengthAbove1000 bool
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
		return 0, fmt.Errorf("getting recipe (to verify that it exists): %w", err)
	}
	existingReviews, err := s.db.GetReviewsForRecipe(ctx, recipeID)
	if err != nil {
		return 0, fmt.Errorf("getting existing reviews: %w", err)
	}
	if slices.ContainsFunc(existingReviews, func(existingReview models.Review) bool {
		return existingReview.OwnerID == userID
	}) {
		return 0, apperrors.ErrAlreadyExists
	}

	comment = normalizeComment(comment)

	if err := validateReviewInput(rating, comment); err != nil {
		return 0, err
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

// UpdateReview returns ErrNotFound if the user has no review for that recipe, and a ReviewValidationError on validation failure.
func (s *ReviewService) UpdateReview(ctx context.Context, recipeID int, userID int, rating int, comment string) error {
	existingReview, err := s.db.GetReviewByRecipeAndUserID(ctx, recipeID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return err
		}
		return fmt.Errorf("getting existing review: %w", err)
	}
	comment = normalizeComment(comment)
	if err := validateReviewInput(rating, comment); err != nil {
		return err
	}
	if err := s.db.UpdateReview(ctx, existingReview.ID, rating, comment); err != nil {
		return err
	}
	return nil
}

// DeleteReview returns ErrNotFound if the user has no review for that recipe.
func (s *ReviewService) DeleteReview(ctx context.Context, recipeID int, userID int) error {
	existingReview, err := s.db.GetReviewByRecipeAndUserID(ctx, recipeID, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return err
		}
		return fmt.Errorf("getting existing review: %w", err)
	}
	if err := s.db.DeleteReviewByID(ctx, existingReview.ID); err != nil {
		return err
	}
	return nil
}

func normalizeComment(comment string) string {
	return strings.TrimSpace(comment)
}

func validateReviewInput(rating int, comment string) error {
	var validationErr ReviewValidationError
	if rating < 1 || rating > 5 {
		validationErr.RatingNotBetween1And5 = true
	}
	if len(comment) > 1000 {
		validationErr.CommentLengthAbove1000 = true
	}
	if validationErr != (ReviewValidationError{}) {
		return validationErr
	}
	return nil
}
