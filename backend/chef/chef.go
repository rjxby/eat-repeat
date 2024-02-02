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
	LoadRecipes(page int, pageSize int, searchTerm string) (result *store.Recipes, err error)
	LoadServings() (result *[]store.ServingV1, err error)
	SaveServing(serving *store.ServingV1) (err error)
	GetServing(id uint) (result *store.ServingV1, err error)
}

func (p RecipeProc) GetRecipes(page int, pageSize int, searchTerm string) (recipes *store.Recipes, err error) {
	recipes, err = p.engine.LoadRecipes(page, pageSize, searchTerm)
	if err != nil {
		return nil, err
	}

	return recipes, nil
}

func (p RecipeProc) GetServings() (servings *[]store.ServingV1, err error) {
	servings, err = p.engine.LoadServings()
	if err != nil {
		return nil, err
	}

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
