package database

import (
	"context"
	"encoding/json"
	"recipe-web-server/internal/models"
)

func (db *DB) CreateRecipe(ctx context.Context, recipe *models.Recipe) (int, error) {
	query := `
		INSERT INTO recipes (title, description, ingredient_sections, instructions, servings, prep_time_seconds, prep_instructions, cook_time_seconds, meal_types, dietary_tags, other_tags, owner_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	ingredient_sections_string, err := json.Marshal(recipe.IngredientSections)
	if err != nil {
		return 0, err
	}
	mealTypesString, err := json.Marshal(recipe.MealTypes)
	if err != nil {
		return 0, err
	}
	dietaryTagsString, err := json.Marshal(recipe.DietaryTags)
	if err != nil {
		return 0, err
	}
	otherTagsString, err := json.Marshal(recipe.OtherTags)
	if err != nil {
		return 0, err
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
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
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
		return err
	}
	mealTypesString, err := json.Marshal(recipe.MealTypes)
	if err != nil {
		return err
	}
	dietaryTagsString, err := json.Marshal(recipe.DietaryTags)
	if err != nil {
		return err
	}
	otherTagsString, err := json.Marshal(recipe.OtherTags)
	if err != nil {
		return err
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
	return err
}

func (db *DB) GetRecipeById(ctx context.Context, id int) (*models.Recipe, error) {
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
		return nil, err
	}

	err = json.Unmarshal(ingredientSectionsString, &recipe.IngredientSections)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(mealTypesString, &recipe.MealTypes)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(dietaryTagsString, &recipe.DietaryTags)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(otherTagsString, &recipe.OtherTags)
	if err != nil {
		return nil, err
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
		return nil, err
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
			return nil, err
		}

		err = json.Unmarshal(mealTypesString, &ro.MealTypes)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(dietaryTagsString, &ro.DietaryTags)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(dietaryTagsString, &ro.DietaryTags)
		if err != nil {
			return nil, err
		}

		overviews = append(overviews, ro)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return overviews, nil
}

func (db *DB) DeleteRecipeById(ctx context.Context, id int) error {
	query := `
		DELETE FROM recipes
		WHERE id = ?
	`

	_, err := db.ExecContext(ctx, query, id)
	return err
}
