package store

import (
	"time"
)

// TODO: add pagination
type Recipes struct {
	Recipes []RecipeV1
}

// TODO: add pagination
type Ingredients struct {
	Ingredients []IngredientV1
}

type Week struct {
	Days   []Day
	Number int
	Year   int
}

type Day struct {
	ID           string
	IsCurrentDay bool
	Title        string
}

type ServingV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	RecipeID uint
	Recipe   RecipeV1 `gorm:"foreignKey:RecipeID"`

	TargetWeekNumber int
	TargetYear       int

	CookedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RecipeV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Title                    string                 `gorm:"type:varchar(255);unique;not null"`
	Description              string                 `gorm:"type:varchar(4000)"`
	Ingredients              []RecipeV1IngredientV1 `gorm:"foreignKey:RecipeV1ID"`
	PreparationTimeInMinutes uint
	CookingTimeInMinutes     uint
	PdfUrl                   string

	RecipeDifficultyID uint
	RecipeDifficulty   RecipeDifficultyV1 `gorm:"foreignKey:RecipeDifficultyID"`

	Rating      float64 `gorm:"default:0.0"`
	RatingCount uint    `gorm:"default:0"`

	Servings []ServingV1 `gorm:"foreignKey:RecipeID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type RecipeDifficultyV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Name string `gorm:"type:varchar(255);unique;not null"`

	Recipes []RecipeV1 `gorm:"foreignKey:RecipeDifficultyID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type IngredientV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Name string `gorm:"type:varchar(255);unique;not null"`

	UnitID uint   `gorm:"not null"`
	Unit   UnitV1 `gorm:"foreignKey:UnitID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UnitV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Name string `gorm:"type:varchar(255);unique;not null"`

	Ingredients []IngredientV1 `gorm:"foreignKey:UnitID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type PantryV1 struct {
	IngredientV1ID uint
	Ingredient     IngredientV1 `gorm:"foreignKey:IngredientV1ID"`

	Amount float64
}

type RecipeV1IngredientV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	RecipeV1ID uint
	Recipe     RecipeV1 `gorm:"foreignKey:RecipeV1ID"`

	IngredientV1ID uint
	Ingredient     IngredientV1 `gorm:"foreignKey:IngredientV1ID"`

	Amount float64
}
