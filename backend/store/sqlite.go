package store

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var databaseName = "data/eatrepeat.sqlite"

// Error messages
var (
	ErrLoadRejected = fmt.Errorf("message expired or deleted")
	ErrSaveRejected = fmt.Errorf("can't save message")
)

type Database struct {
	db *gorm.DB
}

// NewDatabase makes persistent sqlite based store
func NewDatabase() (*Database, error) {
	log.Printf("[INFO] sqlite (persistent) store")
	result := Database{}

	db, err := gorm.Open(sqlite.Open(databaseName), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	result.db = db

	return &result, nil
}

func (s *Database) Migrate() error {
	log.Printf("[INFO] migrating database")

	err := s.db.AutoMigrate(
		&UnitV1{},
		&IngredientV1{},
		&PantryV1{},
		&RecipeV1{},
		&RecipeDifficultyV1{},
		&RecipeV1IngredientV1{},
		&ServingV1{})

	if err != nil {
		return err
	}

	err = Seed(s.db)
	if err != nil {
		return err
	}

	log.Printf("[INFO] database migrated")
	return nil
}

func (s *Database) SaveRecipe(recipe *RecipeV1) (err error) {

	s.db.Create(recipe)

	log.Printf("[DEBUG] saved, recipe=%v", recipe.Title)
	return nil
}

func (s *Database) LoadRecipes() (result *Recipes, err error) {
	log.Printf("[DEBUG] loading all recipes")

	var recipes []RecipeV1
	s.db.Find(&recipes)

	result = &Recipes{Recipes: recipes}

	return result, nil
}

func (s *Database) LoadServingsByWeekYear(weekNumber int, year int) (result *[]ServingV1, err error) {
	log.Printf("[DEBUG] loading servings by week year")

	var servings []ServingV1
	s.db.Where("target_week_number = ? AND target_year = ? AND cooked_at IS NULL", weekNumber, year).Preload("Recipe").Preload("Recipe.RecipeDifficulty").Preload("Recipe.Ingredients").Preload("Recipe.Ingredients.Unit").Find(&servings)

	return &servings, nil
}

func (s *Database) GetIngredients() (result *Ingredients, err error) {
	log.Printf("[DEBUG] loading all ingredients")

	var ingredients []IngredientV1
	s.db.Preload("Unit").Find(&ingredients)

	result = &Ingredients{Ingredients: ingredients}

	return result, nil
}

func (s *Database) GetUnits() (result *[]UnitV1, err error) {
	log.Printf("[DEBUG] loading all units")

	var units []UnitV1
	s.db.Find(&units)

	return &units, nil
}

func (s *Database) SaveIngredient(ingredient *IngredientV1) (err error) {
	log.Printf("[DEBUG] saving ingredient")

	s.db.Create(ingredient)

	return nil
}

func (s *Database) SaveServing(serving *ServingV1) (err error) {
	log.Printf("[DEBUG] saving serving")

	s.db.Save(serving)

	return nil
}

func (s *Database) GetServing(id uint) (result *ServingV1, err error) {
	log.Printf("[DEBUG] loading serving")

	var serving ServingV1
	s.db.Preload("Recipe").Preload("Recipe.RecipeDifficulty").Preload("Recipe.Ingredients").Preload("Recipe.Ingredients.Unit").Where("id = ?", id).First(&serving)

	return &serving, nil
}
