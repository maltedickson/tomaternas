package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/maltedickson/tomaternas/internal/apperrors"
	"github.com/maltedickson/tomaternas/internal/models"
)

func (db *DB) CreateRecipe(ctx context.Context, recipe *models.Recipe) (int, error) {
	query := `
		INSERT INTO recipes (title, description, ingredient_sections, instructions, servings, prep_time_seconds, prep_instructions, cook_time_seconds, meal_types, dietary_tags, other_tags, owner_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	ingredient_sections_string, err := json.Marshal(recipe.IngredientSections)
	if err != nil {
		return 0, fmt.Errorf("marshaling ingredient sections: %w", err)
	}
	mealTypesString, err := json.Marshal(recipe.MealTypes)
	if err != nil {
		return 0, fmt.Errorf("marshaling meal types: %w", err)
	}
	dietaryTagsString, err := json.Marshal(recipe.DietaryTags)
	if err != nil {
		return 0, fmt.Errorf("marshaling dietary tags: %w", err)
	}
	otherTagsString, err := json.Marshal(recipe.OtherTags)
	if err != nil {
		return 0, fmt.Errorf("marshaling other tags: %w", err)
	}

	result, err := db.ExecContext(
		ctx,
		query,
		recipe.Title,
		recipe.Description,
		ingredient_sections_string,
		recipe.Instructions,
		recipe.Servings,
		recipe.PrepTimeSeconds,
		recipe.PrepInstructions,
		recipe.CookTimeSeconds,
		mealTypesString,
		dietaryTagsString,
		otherTagsString,
		recipe.OwnerID,
	)
	if err != nil {
		return 0, fmt.Errorf("db inserting recipe: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("db getting ID for inserted recipe: %w", err)
	}

	return int(id), nil
}

func (db *DB) UpdateRecipe(ctx context.Context, recipe *models.Recipe) error {
	query := `
		UPDATE recipes
		SET title = ?, description = ?, ingredient_sections = ?, instructions = ?, servings = ?, prep_time_seconds = ?, prep_instructions = ?, cook_time_seconds = ?, meal_types = ?, dietary_tags = ?, other_tags = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	ingredientSectionsString, err := json.Marshal(recipe.IngredientSections)
	if err != nil {
		return fmt.Errorf("unmarshaling ingredient sections: %w", err)
	}
	mealTypesString, err := json.Marshal(recipe.MealTypes)
	if err != nil {
		return fmt.Errorf("unmarshaling meal types: %w", err)
	}
	dietaryTagsString, err := json.Marshal(recipe.DietaryTags)
	if err != nil {
		return fmt.Errorf("unmarshaling dietary tags: %w", err)
	}
	otherTagsString, err := json.Marshal(recipe.OtherTags)
	if err != nil {
		return fmt.Errorf("unmarshaling other tags: %w", err)
	}

	_, err = db.ExecContext(
		ctx,
		query,
		recipe.Title,
		recipe.Description,
		ingredientSectionsString,
		recipe.Instructions,
		recipe.Servings,
		recipe.PrepTimeSeconds,
		recipe.PrepInstructions,
		recipe.CookTimeSeconds,
		mealTypesString,
		dietaryTagsString,
		otherTagsString,
		recipe.ID,
	)
	if err != nil {
		return fmt.Errorf("updating recipe %d: %w", recipe.ID, err)
	}
	return nil
}

// GetRecipeByID fetches the recipe with the specified ID from the database.
// If no such recipe exists, GetRecipeByID returns apperrors.ErrNotFound.
func (db *DB) GetRecipeByID(ctx context.Context, id int) (*models.Recipe, error) {
	query := `
		SELECT id, title, description, ingredient_sections, instructions, servings, prep_time_seconds, prep_instructions, cook_time_seconds, meal_types, dietary_tags, other_tags, owner_id, created_at, updated_at
		FROM recipes
		WHERE id = ?
	`

	var recipe models.Recipe
	var ingredientSectionsString []byte
	var mealTypesString []byte
	var dietaryTagsString []byte
	var otherTagsString []byte

	err := db.QueryRowContext(ctx, query, id).Scan(
		&recipe.ID,
		&recipe.Title,
		&recipe.Description,
		&ingredientSectionsString,
		&recipe.Instructions,
		&recipe.Servings,
		&recipe.PrepTimeSeconds,
		&recipe.PrepInstructions,
		&recipe.CookTimeSeconds,
		&mealTypesString,
		&dietaryTagsString,
		&otherTagsString,
		&recipe.OwnerID,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("db getting recipe %d: %w", id, err)
	}

	err = json.Unmarshal(ingredientSectionsString, &recipe.IngredientSections)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling ingredient sections: %w", err)
	}

	err = json.Unmarshal(mealTypesString, &recipe.MealTypes)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling meal types: %w", err)
	}

	err = json.Unmarshal(dietaryTagsString, &recipe.DietaryTags)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling dietary tags: %w", err)
	}

	err = json.Unmarshal(otherTagsString, &recipe.OtherTags)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling other tags: %w", err)
	}

	return &recipe, nil
}

func (db *DB) GetAllRecipeOverviews(ctx context.Context) ([]models.RecipeOverview, error) {
	query := `
		SELECT id, title, prep_time_seconds, cook_time_seconds, meal_types, dietary_tags, other_tags, owner_id, created_at, updated_at
		FROM recipes
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("db selecting recipe overviews: %w", err)
	}
	defer rows.Close()

	var overviews []models.RecipeOverview

	for rows.Next() {
		var ro models.RecipeOverview
		var mealTypesString, dietaryTagsString, otherTagsString []byte

		err := rows.Scan(
			&ro.ID,
			&ro.Title,
			&ro.PrepTimeSeconds,
			&ro.CookTimeSeconds,
			&mealTypesString,
			&dietaryTagsString,
			&otherTagsString,
			&ro.OwnerID,
			&ro.CreatedAt,
			&ro.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("db scanning recipe overview: %w", err)
		}

		err = json.Unmarshal(mealTypesString, &ro.MealTypes)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling meal types: %w", err)
		}

		err = json.Unmarshal(dietaryTagsString, &ro.DietaryTags)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling dietary tags: %w", err)
		}

		err = json.Unmarshal(dietaryTagsString, &ro.DietaryTags)
		if err != nil {
			return nil, fmt.Errorf("unmarshaling dietary tags: %w", err)
		}

		overviews = append(overviews, ro)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("db iterating recipes: %w", err)
	}

	return overviews, nil
}

func (db *DB) DeleteRecipeByID(ctx context.Context, id int) error {
	query := `
		DELETE FROM recipes
		WHERE id = ?
	`
	_, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("db deleting recipe %d: %w", id, err)
	}
	return nil
}
