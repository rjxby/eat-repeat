package store

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

var pdfStore = "data/recipes/"
var imageStore = "data/images/"

func Seed(db *gorm.DB) error {
	// read csv file
	file, err := os.Open("backend/store/seed-data/data.csv")
	if err != nil {
		log.Fatalf("[ERROR] Failed to open CSV file: %v", err)
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("[ERROR] Failed to parse CSV file: %v", err)
		return err
	}

	// skip header
	records = records[1:]

	// group records from file by recipe title (first column)
	groupedRecords := make(map[string][][]string)
	for _, record := range records {
		title := record[0]
		groupedRecords[title] = append(groupedRecords[title], record)
	}

	// get existing recipes
	var existingRecipes []RecipeV1
	if result := db.Find(&existingRecipes); result.Error != nil {
		log.Fatalf("[ERROR] getting recipes data: %v\n", result.Error)
		return result.Error
	}

	// get existing ingredients
	var existingIngredients []IngredientV1
	if result := db.Find(&existingIngredients); result.Error != nil {
		log.Fatalf("[ERROR] getting ingredients data: %v\n", result.Error)
		return result.Error
	}

	// get existing units
	var existingUnits []UnitV1
	if result := db.Find(&existingUnits); result.Error != nil {
		log.Fatalf("[ERROR] getting units data: %v\n", result.Error)
		return result.Error
	}

	unitsToSeed := []UnitV1{}
	ingridientsToSeed := []IngredientV1{}
	recipesToSeed := []RecipeV1{}

	// seed data from csv file if not exists
	for title, records := range groupedRecords {
		if containsRecipe(existingRecipes, title) {
			continue
		}

		for _, record := range records {
			unit := UnitV1{Name: record[2]}
			if !containsUnit(existingUnits, unit.Name) && !containsUnit(unitsToSeed, unit.Name) {
				unitsToSeed = append(unitsToSeed, unit)
			}

			ingridient := IngredientV1{Name: record[3], Unit: unit}
			if !containsIngredient(existingIngredients, ingridient.Name) && !containsIngredient(ingridientsToSeed, ingridient.Name) {
				ingridientsToSeed = append(ingridientsToSeed, ingridient)
			}
		}

		recipePdfUrl := buildRecipePdfUrl(title)
		recipeThumbnailUrl := buildRecipeThumbnailUrl(title)

		recipesToSeed = append(recipesToSeed, RecipeV1{Title: title, Description: title, PdfUrl: recipePdfUrl, ThumbnailUrl: recipeThumbnailUrl})
	}

	err = seedUnits(db, unitsToSeed)
	if err != nil {
		log.Fatalf("[ERROR] seeding units data: %v\n", err)
		return err
	}
	// fetch units from db to get the new ids
	if result := db.Find(&existingUnits); result.Error != nil {
		log.Fatalf("[ERROR] getting units data: %v\n", result.Error)
		return result.Error
	}

	// update ingredient units with new ids
	for i := range ingridientsToSeed {
		// find the unit in existing units
		for _, existingUnit := range existingUnits {
			if existingUnit.Name == ingridientsToSeed[i].Unit.Name {
				ingridientsToSeed[i].UnitID = existingUnit.ID
				ingridientsToSeed[i].Unit = existingUnit
				break
			}
		}
	}
	err = seedIngredients(db, ingridientsToSeed)
	if err != nil {
		log.Fatalf("[ERROR] seeding ingredients data: %v\n", err)
		return err
	}
	// fetch ingredients from db to get the new ids
	if result := db.Find(&existingIngredients); result.Error != nil {
		log.Fatalf("[ERROR] getting ingredients data: %v\n", result.Error)
		return result.Error
	}

	// update recipe ingredients with new ids
	for i := range recipesToSeed {
		// find the ingridient in existing recipes
		currentRecipe := recipesToSeed[i]
		for _, existingIngredient := range existingIngredients {
			if existingIngredient.Name == currentRecipe.Title {
				currentRecipe.ID = existingIngredient.ID
				break
			}
		}
	}
	err = seedRecipes(db, recipesToSeed)
	if err != nil {
		log.Fatalf("[ERROR] seeding recipes data: %v\n", err)
		return err
	}
	// fetch recipes from db to get the new ids
	if result := db.Find(&existingRecipes); result.Error != nil {
		log.Fatalf("[ERROR] getting recipes data: %v\n", result.Error)
		return result.Error
	}

	// seed recipe related ingredients
	recipeIngredientToSeed := []RecipeV1IngredientV1{}
	for title, records := range groupedRecords {
		for _, record := range records {
			recipe := RecipeV1{}
			for _, existingRecipe := range existingRecipes {
				if existingRecipe.Title == title {
					recipe = existingRecipe
					break
				}
			}

			ingredient := IngredientV1{}
			for _, existingIngredient := range existingIngredients {
				if existingIngredient.Name == record[3] {
					ingredient = existingIngredient
					break
				}
			}

			ingridientAmount, err := strconv.ParseFloat(record[1], 64)
			if err != nil {
				log.Printf("[ERROR] parsing unit amount for recipe %v %v", record[3], err)
				continue
			}

			recipeIngredient := RecipeV1IngredientV1{RecipeV1ID: recipe.ID, IngredientV1ID: ingredient.ID, Amount: float64(ingridientAmount)}
			recipeIngredientToSeed = append(recipeIngredientToSeed, recipeIngredient)
		}
	}

	err = seedRecipeIngredients(db, recipeIngredientToSeed)
	if err != nil {
		log.Fatalf("[ERROR] seeding recipe ingredients data: %v\n", err)
		return err
	}

	return nil
}

func seedUnits(db *gorm.DB, units []UnitV1) error {
	for _, unit := range units {
		if result := db.Create(&unit); result.Error != nil {
			log.Fatalf("[ERROR] seeding units data: %v\n", result.Error)
			return result.Error
		}
	}

	return nil
}

func seedIngredients(db *gorm.DB, ingredients []IngredientV1) error {
	for _, ingredient := range ingredients {
		if result := db.Create(&ingredient); result.Error != nil {
			log.Fatalf("[ERROR] seeding ingredients data: %v\n", result.Error)
			return result.Error
		}
	}

	return nil
}

func buildRecipePdfUrl(title string) sql.NullString {
	lowerTitle := strings.ToLower(title)
	recipePdfUrl := fmt.Sprintf("%s%s.pdf", pdfStore, lowerTitle)
	if _, err := os.Stat(recipePdfUrl); errors.Is(err, os.ErrNotExist) {
		return sql.NullString{}
	}

	return sql.NullString{
		String: recipePdfUrl,
		Valid:  true,
	}
}

func buildRecipeThumbnailUrl(title string) sql.NullString {
	lowerTitle := strings.ToLower(title)
	recipeThumbnailUrl := fmt.Sprintf("%s%s/1.jpeg", imageStore, lowerTitle)
	if _, err := os.Stat(recipeThumbnailUrl); errors.Is(err, os.ErrNotExist) {
		return sql.NullString{}
	}

	return sql.NullString{
		String: recipeThumbnailUrl,
		Valid:  true,
	}
}

func seedRecipes(db *gorm.DB, recipes []RecipeV1) error {
	for _, recipe := range recipes {
		if result := db.Create(&recipe); result.Error != nil {
			log.Fatalf("[ERROR] seeding recipes data: %v\n", result.Error)
			return result.Error
		}
	}

	return nil
}

func seedRecipeIngredients(db *gorm.DB, recipeIngredients []RecipeV1IngredientV1) error {
	for _, recipeIngredient := range recipeIngredients {
		if result := db.Create(&recipeIngredient); result.Error != nil {
			log.Fatalf("[ERROR] seeding recipe ingredients data: %v\n", result.Error)
			return result.Error
		}
	}

	return nil
}

func containsRecipe(recipes []RecipeV1, recipeTitle string) bool {
	for _, r := range recipes {
		if r.Title == recipeTitle {
			return true
		}
	}

	return false
}

func containsIngredient(ingredients []IngredientV1, ingredientName string) bool {
	for _, i := range ingredients {
		if i.Name == ingredientName {
			return true
		}
	}

	return false
}

func containsUnit(units []UnitV1, unitName string) bool {
	for _, u := range units {
		if u.Name == unitName {
			return true
		}
	}

	return false
}
