package database

import (
	"fmt"
	"recipe-web-server/internal/models"
)

func (db *DB) CreateReview(review *models.Review) (int, error) {
	query := `
		INSERT INTO reviews (recipe_id, owner_id, rating, comment)
		VALUES (?, ?, ?, ?)
	`

	result, err := db.Exec(
		query,
		review.RecipeID,
		review.OwnerID,
		review.Rating,
		review.Comment,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting review into database: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("retrieving ID of inserted review: %w", err)
	}

	return int(id), nil
}

func (db *DB) UpdateReview(review *models.Review) error {
	query := `
		UPDATE reviews
		SET rating = ?, comment = ?
		WHERE id = ?
	`
	_, err := db.Exec(
		query,
		review.Rating,
		review.Comment,
		review.ID,
	)
	if err != nil {
		return fmt.Errorf("executing query to update review with ID %d: %w", review.ID, err)
	}
	return nil
}

func (db *DB) GetReviewById(id int) (*models.Review, error) {
	query := `
		SELECT id, recipe_id, owner_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE id = ?
	`
	var review models.Review
	err := db.QueryRow(query, id).Scan(
		&review.ID,
		&review.RecipeID,
		&review.OwnerID,
		&review.Rating,
		&review.Comment,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("fetching review with id %d: %w", id, err)
	}
	return &review, nil
}

func (db *DB) GetReviewsForRecipe(recipeID int) ([]models.Review, error) {
	query := `
		SELECT id, recipe_id, owner_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE recipe_id = ?
	`

	rows, err := db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("fetching reviews for recipe with ID %d from database: %w", recipeID, err)
	}
	defer rows.Close()
	var reviews []models.Review
	for rows.Next() {
		var review models.Review

		err := rows.Scan(
			&review.ID,
			&review.RecipeID,
			&review.OwnerID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning database row: %w", err)
		}
		reviews = append(reviews, review)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows fetched from database: %w", err)
	}
	return reviews, nil
}

func (db *DB) DeleteReviewById(id int) error {
	query := `
		DELETE FROM reviews
		WHERE id = ?
	`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("executing query to delete review with ID %d from database: %w", id, err)
	}
	return nil
}
