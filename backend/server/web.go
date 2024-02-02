package server

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rjxby/eat-repeat/backend/store"
	"github.com/rjxby/eat-repeat/frontend"
)

const (
	baseTmpl = "base"

	servingsTmplName       = "servings.tmpl.html"
	recipesTmplName        = "recipes.tmpl.html"
	moreRecipesTmplName    = "more-recipes.tmpl.html"
	pantryTmplName         = "pantry.tmpl.html"
	ingredientFormTmplName = "ingredient-form.tmpl.html"
)

type servingsView struct {
	Servings []store.ServingV1
}

type recipesView struct {
	RecipesCards recipesCardsView
}

type recipesCardsView struct {
	Recipes    []store.RecipeV1
	Page       int
	PageSize   int
	SearchTerm string
}

type pantryView struct {
	Ingredients []store.IngredientV1
}

type pantryFormView struct {
	Units []store.UnitV1
}

type templateData struct {
	View any
}

// render renders a template
func (s Server) render(w http.ResponseWriter, status int, page, tmplName string, data any) {
	ts, ok := s.TemplateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	if tmplName == "" {
		tmplName = baseTmpl
	}

	log.Printf("[DEBUG] rendering %s", tmplName)

	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	log.Printf("[DEBUG] rendered %s", tmplName)

	w.WriteHeader(status)
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// renders the home page with servings
// GET /
func (s Server) indexCtrl(w http.ResponseWriter, r *http.Request) {
	servings, err := s.Chef.GetServings()
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := templateData{
		View: servingsView{
			Servings: *servings,
		},
	}

	s.render(w, http.StatusOK, servingsTmplName, baseTmpl, data)
}

// mark a serving as cooked
// POST /servings/cooked
func (s Server) cookedViewCtrl(w http.ResponseWriter, r *http.Request) {
	servingId, err := strconv.ParseUint(r.FormValue("servingID"), 10, 32)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	serving, err := s.Chef.GetServing(uint(servingId))
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	serving.CookedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	if err := s.Chef.SaveServing(serving); err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// renders the show recipes page
// GET /recipes
func (s Server) recipesViewCtrl(w http.ResponseWriter, r *http.Request) {
	// it's 9 elements page size due to grid size on HTML, search by default is empty string
	recipes, err := s.Chef.GetRecipes(1, 9, "")
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := templateData{
		View: recipesView{
			RecipesCards: recipesCardsView{
				Recipes:    recipes.Recipes,
				Page:       recipes.Page,
				PageSize:   recipes.PageSize,
				SearchTerm: recipes.SearchTerm,
			},
		},
	}

	s.render(w, http.StatusOK, recipesTmplName, recipesTmplName, data)
}

// renders the show more recipes cards
// GET /recipes/more
func (s Server) moreRecipesViewCtrl(w http.ResponseWriter, r *http.Request) {

	// Parse the page and pageSize from the query parameters
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(w, "invalid page parameter", http.StatusBadRequest)
		return
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		http.Error(w, "invalid pageSize parameter", http.StatusBadRequest)
		return
	}

	searchTerm := r.URL.Query().Get("searchTerm")

	recipes, err := s.Chef.GetRecipes(page, pageSize, searchTerm)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := templateData{
		View: recipesCardsView{
			Recipes:    recipes.Recipes,
			Page:       recipes.Page,
			PageSize:   recipes.PageSize,
			SearchTerm: recipes.SearchTerm,
		},
	}

	s.render(w, http.StatusOK, moreRecipesTmplName, moreRecipesTmplName, data)
}

// re-renders the recipes page with a mutaded recipe
// POST /recipes/select
func (s Server) selectRecipeViewCtrl(w http.ResponseWriter, r *http.Request) {
	selectedRecipeId, err := strconv.ParseUint(r.FormValue("recipeID"), 10, 32)
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	serving := store.ServingV1{
		RecipeID: uint(selectedRecipeId),
	}

	if err := s.Chef.SaveServing(&serving); err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

// renders the show pantry page
// GET /pantry
func (s Server) pantryViewCtrl(w http.ResponseWriter, r *http.Request) {
	ingridients, err := s.Pantry.GetIngredients()

	log.Printf("[DEBUG] ingridients: %v", ingridients)

	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := templateData{
		View: pantryView{
			Ingredients: ingridients.Ingredients,
		},
	}

	s.render(w, http.StatusOK, pantryTmplName, pantryTmplName, data)
}

// renders the ingridient form
// GET /pantry/add
func (s Server) ingredientFormViewCtrl(w http.ResponseWriter, r *http.Request) {
	units, err := s.Pantry.GetUnits()
	if err != nil {
		log.Printf("[ERROR] %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data := templateData{
		View: pantryFormView{
			Units: *units,
		},
	}

	s.render(w, http.StatusOK, ingredientFormTmplName, ingredientFormTmplName, data)
}

// TODO version 0.0.2
// create a new ingredient and redirect to the pantry page
// POST /pantry
// func (s Server) createIngredientViewCtrl(w http.ResponseWriter, r *http.Request) {
// 	unitId, err := strconv.ParseUint(r.FormValue("unitID"), 10, 32)
// 	if err != nil {
// 		log.Printf("[ERROR] %v", err)
// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// 		return
// 	}

// 	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
// 	if err != nil {
// 		log.Printf("[ERROR] %v", err)
// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// 		return
// 	}

// 	ingredient := store.IngredientV1{
// 		Name:   r.FormValue("name"),
// 		Amount: amount,
// 		UnitID: uint(unitId),
// 	}

// 	if err := s.Pantry.SaveIngredient(&ingredient); err != nil {
// 		log.Printf("[ERROR] %v", err)
// 		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
// 		return
// 	}

// 	http.Redirect(w, r, "/pantry", http.StatusSeeOther)
// }

// renders ingredient form with ingredient data
// GET /pantry/edit/{id}
func (s Server) editIngredientViewCtrl(w http.ResponseWriter, r *http.Request) {
	// TODO version 0.0.2
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(frontend.Templates, "html/*/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/index.tmpl.html",
			"html/sections/*.tmpl.html",
			page,
		}

		ts, err := template.New(name).Funcs(template.FuncMap{"until": until, "subtract": subtract, "add": add, "toLowerStr": toLowerStr}).ParseFS(frontend.Templates, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}

// until is a helper function for templates to generate a slice of numbers
func until(n int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = i
	}
	return result
}

func subtract(first int, second int) int {
	return first - second
}

func add(first int, second int) int {
	return first + second
}

func toLowerStr(input string) string {
	return strings.ToLower(input)
}
