package models

import (
	"time"
)

const (
	TagVegetarian = "Vegetarisk"
	TagVegan      = "Vegansk"
	TagMilkFree   = "Mjölkfri"
)

var AllowedMealTypes = []string{"Frukost", "Förrätt", "Huvudrätt", "Tillbehör", "Fika/efterrätt", "Dryck"}
var AllowedDietaryTags = []string{TagVegan, TagVegetarian, "Glutenfri", TagMilkFree}
var AllowedOtherTags = []string{"Festlig", "Matlåda", "Storkok"}

type Recipe struct {
	ID                 int
	Title              string
	Description        string
	IngredientSections []IngredientSection
	Instructions       string
	Servings           string
	PrepTimeSeconds    int
	PrepInstructions   string
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

type RecipeOverview struct {
	ID              int
	Title           string
	PrepTimeSeconds int
	CookTimeSeconds int
	MealTypes       []string
	DietaryTags     []string
	OtherTags       []string
	OwnerID         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
