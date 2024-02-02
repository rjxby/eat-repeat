package store

import (
	"database/sql"
	"time"
)

type Recipes struct {
	Recipes    []RecipeV1
	Page       int
	PageSize   int
	SearchTerm string
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

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusInProgress JobStatus = "in_progress"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type JobV1 struct {
	ID     uint      `gorm:"primaryKey;autoIncrement"`
	Status JobStatus `gorm:"not null"`

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type ServingV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	RecipeID uint
	Recipe   RecipeV1 `gorm:"foreignKey:RecipeID"`

	CookedAt sql.NullTime

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type RecipeV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Title                    string                 `gorm:"type:varchar(255);unique;not null"`
	Description              string                 `gorm:"type:varchar(4000)"`
	Ingredients              []RecipeV1IngredientV1 `gorm:"foreignKey:RecipeV1ID"`
	PreparationTimeInMinutes uint
	CookingTimeInMinutes     uint
	ThumbnailUrl             sql.NullString
	PdfUrl                   sql.NullString

	RecipeDifficultyID uint
	RecipeDifficulty   RecipeDifficultyV1 `gorm:"foreignKey:RecipeDifficultyID"`

	Rating      float64 `gorm:"default:0.0"`
	RatingCount uint    `gorm:"default:0"`

	Servings []ServingV1 `gorm:"foreignKey:RecipeID"`

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type RecipeDifficultyV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Name string `gorm:"type:varchar(255);unique;not null"`

	Recipes []RecipeV1 `gorm:"foreignKey:RecipeDifficultyID"`

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type IngredientV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Name string `gorm:"type:varchar(255);unique;not null"`

	UnitID uint   `gorm:"not null"`
	Unit   UnitV1 `gorm:"foreignKey:UnitID"`

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type UnitV1 struct {
	ID uint `gorm:"primaryKey;autoIncrement"`

	Name string `gorm:"type:varchar(255);unique;not null"`

	Ingredients []IngredientV1 `gorm:"foreignKey:UnitID"`

	CreatedAt time.Time
	UpdatedAt sql.NullTime
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
