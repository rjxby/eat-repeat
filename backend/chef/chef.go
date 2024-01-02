package chef

import (
	"log"

	"github.com/rjxby/eat-repeat/backend/store"
)

// RecipeProc creates and save recipes
type RecipeProc struct {
	engine Engine
}

// New makes RecipeProc
func New(engine Engine) *RecipeProc {
	return &RecipeProc{
		engine: engine,
	}
}

// Engine defines interface to save and load recipes
type Engine interface {
	LoadRecipes() (result *store.Recipes, err error)
	LoadServingsByWeekYear(weekNumber int, year int) (result *[]store.ServingV1, err error)
	SaveServing(serving *store.ServingV1) (err error)
	GetServing(id uint) (result *store.ServingV1, err error)
}

func (p RecipeProc) GetRecipes() (recipes *store.Recipes, err error) {
	recipes, err = p.engine.LoadRecipes()
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] recipes are loaded: %v", recipes)

	return recipes, nil
}

func (p RecipeProc) GetServingsByWeekYear(weekNumber int, year int) (servings *[]store.ServingV1, err error) {
	servings, err = p.engine.LoadServingsByWeekYear(weekNumber, year)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] servings are loaded for %v-%v: %v", weekNumber, year, servings)

	return servings, nil
}

func (p RecipeProc) SaveServing(serving *store.ServingV1) (err error) {
	err = p.engine.SaveServing(serving)
	if err != nil {
		return err
	}

	log.Printf("[INFO] serving is saved: %v", serving)

	return nil
}

func (p RecipeProc) GetServing(id uint) (serving *store.ServingV1, err error) {
	serving, err = p.engine.GetServing(id)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] serving is loaded: %v", serving)

	return serving, nil
}
