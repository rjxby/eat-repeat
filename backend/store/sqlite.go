package store

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var databaseName = "data/eatrepeat.sqlite"

type Database struct {
	db *gorm.DB
}

// NewDatabase makes persistent sqlite based store
func NewDatabase() (*Database, error) {
	log.Printf("[INFO] sqlite (persistent) store")
	result := Database{}

	db, err := gorm.Open(sqlite.Open(databaseName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	result.db = db

	return &result, nil
}

func (s *Database) Migrate() error {
	log.Printf("[INFO] migrating database")

	if err := s.db.AutoMigrate(
		&JobV1{},
		&UnitV1{},
		&IngredientV1{},
		&PantryV1{},
		&RecipeV1{},
		&RecipeDifficultyV1{},
		&RecipeV1IngredientV1{},
		&ServingV1{}); err != nil {
		return err
	}

	if err := s.addFullTextSearch(); err != nil {
		return err
	}

	if err := Seed(s.db); err != nil {
		return err
	}

	log.Printf("[INFO] database migrated")
	return nil
}

func (s *Database) addFullTextSearch() (err error) {
	// enable full text search (fts) for recipes (table name is 'recipe_v1' by convention)
	// create fts table and triggers to sync the data between fts table and main table
	s.db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS recipe_v1_fts USING fts5(title);

		CREATE TRIGGER IF NOT EXISTS recipe_v1_ai AFTER INSERT ON recipe_v1
		BEGIN
			INSERT INTO recipe_v1_fts (rowid, title)
			VALUES (new.id, new.title);
		END;

		CREATE TRIGGER IF NOT EXISTS recipe_v1_ad AFTER DELETE ON recipe_v1
		BEGIN
			DELETE FROM recipe_v1_fts WHERE title = old.title;
		END;

		CREATE TRIGGER IF NOT EXISTS recipe_v1_au AFTER UPDATE ON recipe_v1
		BEGIN
			DELETE FROM recipe_v1_fts WHERE title = old.title;
			INSERT INTO recipe_v1_fts (rowid, title)
			VALUES (new.id, new.title);
		END;
	`)

	return nil
}

func (s *Database) SaveRecipe(recipe *RecipeV1) (savedRecipe *RecipeV1, err error) {
	s.db.Save(recipe)
	return recipe, nil
}

func (s *Database) SaveSyncJob(job *JobV1) (savedJob *JobV1, err error) {
	s.db.Save(job)
	return job, nil
}

func (s *Database) LoadRecipes(page int, pageSize int, searchTerm string) (result *Recipes, err error) {
	var recipes []RecipeV1
	offset := (page - 1) * pageSize

	// Use a join between RecipeV1 and recipe_v1_fts using the id column
	// Perform a full-text search on the title column of recipe_v1_fts if searchTerm is provided
	query := s.db.Table("recipe_v1_fts").
		Select("recipe_v1.*").
		Joins("JOIN recipe_v1 ON recipe_v1_fts.rowid = recipe_v1.id").
		Preload("Ingredients.Ingredient").
		Preload("Ingredients.Ingredient.Unit")

	// Conditionally add the WHERE clause only if searchTerm is provided
	if searchTerm != "" {
		query = query.Where("recipe_v1_fts.title MATCH ? ORDER BY recipe_v1_fts.rank", searchTerm)
	}

	query.Offset(offset).Limit(pageSize).Find(&recipes)

	result = &Recipes{Recipes: recipes, Page: page, PageSize: pageSize, SearchTerm: searchTerm}

	return result, nil
}

func (s *Database) GetRecipe(id uint) (result *RecipeV1, err error) {
	var recipe RecipeV1
	s.db.Where("id = ?", id).First(&recipe)

	return &recipe, nil
}

func (s *Database) LoadServings() (result *[]ServingV1, err error) {
	var servings []ServingV1
	s.db.Where("cooked_at IS NULL").Preload("Recipe").Preload("Recipe.RecipeDifficulty").Preload("Recipe.Ingredients").Preload("Recipe.Ingredients.Ingredient").Preload("Recipe.Ingredients.Ingredient.Unit").Find(&servings)

	return &servings, nil
}

func (s *Database) GetIngredients() (result *Ingredients, err error) {
	var ingredients []IngredientV1
	s.db.Preload("Unit").Find(&ingredients)

	result = &Ingredients{Ingredients: ingredients}

	return result, nil
}

func (s *Database) GetUnits() (result *[]UnitV1, err error) {
	var units []UnitV1
	s.db.Find(&units)

	return &units, nil
}

func (s *Database) SaveIngredient(ingredient *IngredientV1) (err error) {
	s.db.Save(ingredient)
	return nil
}

func (s *Database) SaveServing(serving *ServingV1) (err error) {
	s.db.Save(serving)
	return nil
}

func (s *Database) GetServing(id uint) (result *ServingV1, err error) {
	var serving ServingV1
	s.db.Preload("Recipe").Preload("Recipe.RecipeDifficulty").Preload("Recipe.Ingredients").Preload("Recipe.Ingredients.Unit").Where("id = ?", id).First(&serving)

	return &serving, nil
}
