package database

import (
	"encoding/json"
	"recipe-web-server/internal/models"
)

func (db *DB) CreateRecipe(recipe *models.Recipe) (int, error) {
	query := `
		INSERT INTO recipes (title, description, ingredient_sections, instructions, servings, prep_time_seconds, cook_time_seconds, meal_types, dietary_tags, other_tags, owner_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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

	result, err := db.Exec(
		query,
		recipe.Title,
		recipe.Description,
		ingredient_sections_string,
		recipe.Instructions,
		recipe.Servings,
		recipe.PrepTimeSeconds,
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

func (db *DB) GetRecipeById(id int) (*models.Recipe, error) {
	query := `
		SELECT id, title, description, ingredient_sections, instructions, servings, prep_time_seconds, cook_time_seconds, meal_types, dietary_tags, other_tags, owner_id, created_at, updated_at
		FROM recipes
		WHERE id = ?
	`

	var recipe models.Recipe
	var ingredientSectionsString []byte
	var mealTypesString []byte
	var dietaryTagsString []byte
	var otherTagsString []byte

	err := db.QueryRow(query, id).Scan(
		&recipe.ID,
		&recipe.Title,
		&recipe.Description,
		&ingredientSectionsString,
		&recipe.Instructions,
		&recipe.Servings,
		&recipe.PrepTimeSeconds,
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

	err = json.Unmarshal(dietaryTagsString, &recipe.DietaryTags)
	if err != nil {
		return nil, err
	}

	return &recipe, nil
}

func (db *DB) DeleteRecipeById(id int) error {
	query := `
		DELETE FROM recipes
		WHERE id = ?
	`

	_, err := db.Exec(query, id)
	return err
}
