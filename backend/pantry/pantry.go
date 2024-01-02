package pantry

import (
	"log"

	"github.com/rjxby/eat-repeat/backend/store"
)

// PantryProc creates and save recipes
type PantryProc struct {
	engine Engine
}

// New makes PantryProc
func New(engine Engine) *PantryProc {
	return &PantryProc{
		engine: engine,
	}
}

// Engine defines interface to save and load ingredients
type Engine interface {
	GetIngredients() (result *store.Ingredients, err error)
	GetUnits() (result *[]store.UnitV1, err error)
	SaveIngredient(ingredient *store.IngredientV1) (err error)
}

func (p PantryProc) GetIngredients() (ingredients *store.Ingredients, err error) {
	ingredients, err = p.engine.GetIngredients()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] ingredients are loaded: %v", ingredients)

	return ingredients, nil
}

func (p PantryProc) GetUnits() (units *[]store.UnitV1, err error) {
	units, err = p.engine.GetUnits()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] units are loaded: %v", units)

	return units, nil
}

func (p PantryProc) SaveIngredient(ingredient *store.IngredientV1) (err error) {
	log.Printf("[INFO] saving ingredient: %v", ingredient)

	err = p.engine.SaveIngredient(ingredient)
	if err != nil {
		return err
	}

	log.Printf("[INFO] ingredient is saved: %v", ingredient)

	return nil
}
