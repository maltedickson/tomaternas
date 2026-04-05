package services

import (
	"recipe-web-server/internal/database"
	"recipe-web-server/internal/models"
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

func (s *RecipeService) GetRecipeById(id int) (*models.Recipe, error) {
	return s.db.GetRecipeById(id)
}
