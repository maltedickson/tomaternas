package models

import (
	"time"
)

type Recipe struct {
	ID                 int
	Title              string
	Description        string
	IngredientSections []IngredientSection
	Instructions       string
	Servings           string
	PrepTimeSeconds    int
	CookTimeSeconds    int
	MealTypes          []string
	DietaryTags        []string
	OtherTags          []string
	OwnerID            int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type IngredientSection struct {
	Heading     string
	Ingredients []Ingredient
}

type Ingredient struct {
	Name   string
	Amount string
}
