package server

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/rjxby/eat-repeat/backend/store"
)

type RecipesResultsJSON struct {
	Page       int          `json:"page"`
	PageSize   int          `json:"pageSize"`
	SearchTerm *string      `json:"searchTerm,omitempty"`
	Recipes    []RecipeJSON `json:"recipes"`
}

type RecipeJSON struct {
	ID                   int      `json:"id"`
	Title                string   `json:"title"`
	Description          string   `json:"description,omitempty"`
	Ingredients          []string `json:"ingredients"`
	CookingTimeInMinutes *int     `json:"cookingTimeInMinutes,omitempty"`
	ThumbnailUrl         *string  `json:"thumbnailUrl,omitempty"`
	PdfUrl               *string  `json:"pdfUrl,omitempty"`
}

// POST /v1/recepies/sync
func (s Server) syncRecepiesCtrl(w http.ResponseWriter, r *http.Request) {

	job, err := s.Worker.RunSyncRecipes()
	if err != nil {
		renderInternalServerError(w, r, "failed to run sync job", err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, job)
}

// GET /v1/recepies
func (s Server) getRecepiesCtrl(w http.ResponseWriter, r *http.Request) {

	// Parse the page and pageSize from the query parameters
	page, err := parseQueryParam(r.URL.Query().Get("page"))
	if err != nil {
		renderBadRequest(w, r, "invalid page parameter", err)
		return
	}

	pageSize, err := parseQueryParam(r.URL.Query().Get("pageSize"))
	if err != nil {
		renderBadRequest(w, r, "invalid pageSize parameter", err)
		return
	}

	searchTerm := strings.TrimSpace(r.URL.Query().Get("searchTerm"))

	recipes, err := s.Chef.GetRecipes(page, pageSize, searchTerm)
	if err != nil {
		renderInternalServerError(w, r, "failed to load recepies", err)
		return
	}

	recipesResults := mapToJSON(s.Settings.StaticContentEndpoint, recipes)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, recipesResults)
}

func mapToJSON(staticContentEndpoint string, recipes *store.Recipes) *RecipesResultsJSON {
	var mappedRecipes []RecipeJSON
	for _, recipe := range recipes.Recipes {
		ingredients := mapIngredients(recipe.Ingredients)
		cookingTimeInMinutes := mapCookingTime(recipe.CookingTimeInMinutes)
		thumbnailUrl := mapOptionalURL(staticContentEndpoint, recipe.ThumbnailUrl)
		pdfUrl := mapOptionalURL(staticContentEndpoint, recipe.PdfUrl)

		mappedRecipes = append(mappedRecipes, RecipeJSON{
			ID:                   int(recipe.ID),
			Title:                recipe.Title,
			Description:          recipe.Description,
			Ingredients:          ingredients,
			CookingTimeInMinutes: cookingTimeInMinutes,
			ThumbnailUrl:         thumbnailUrl,
			PdfUrl:               pdfUrl,
		})
	}

	return &RecipesResultsJSON{
		Page:       recipes.Page,
		PageSize:   recipes.PageSize,
		SearchTerm: &recipes.SearchTerm,
		Recipes:    mappedRecipes,
	}
}

func mapIngredients(src []store.RecipeV1IngredientV1) []string {
	var result []string
	for _, ingredient := range src {
		result = append(result, ingredient.Ingredient.Name)
	}
	return result
}

func mapCookingTime(src uint) *int {
	if src != 0 {
		val := int(src)
		return &val
	}
	return nil
}

func mapOptionalURL(staticContentEndpoint string, src sql.NullString) *string {
	if src.Valid {
		url := staticContentEndpoint + src.String
		return &url
	}
	return nil
}
