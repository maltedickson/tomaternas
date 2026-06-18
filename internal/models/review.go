package models

import "time"

type Review struct {
	ID        int
	RecipeID  int
	OwnerID   int
	Rating    int
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RecipeReviewItem struct {
	Rating          int
	Comment         string
	UpdatedAt       time.Time
	UserDisplayName string
}
