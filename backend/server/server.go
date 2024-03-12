package server

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/rjxby/eat-repeat/backend/store"
	"github.com/rjxby/eat-repeat/frontend"
)

type Server struct {
	Chef          Chef
	Scheduler     Scheduler
	Pantry        Pantry
	Worker        Worker
	Version       string
	TemplateCache map[string]*template.Template
	Settings      Settings
}

type Chef interface {
	GetRecipes(page int, pageSize int, searchTerm string) (recipes *store.Recipes, err error)
	GetServings() (servings *[]store.ServingV1, err error)
	SaveServing(serving *store.ServingV1) (err error)
	GetServing(id uint) (serving *store.ServingV1, err error)
}

type Pantry interface {
	GetIngredients() (ingridients *store.Ingredients, err error)
	GetUnits() (units *[]store.UnitV1, err error)
	SaveIngredient(ingredient *store.IngredientV1) (err error)
}

type Scheduler interface {
	GetWeek() (week *store.Week, err error)
	GetNextWeek() (week *store.Week, err error)
}

type Worker interface {
	RunSyncRecipes() (job *store.JobV1, err error)
}

type Settings struct {
	RunMigration           bool
	PdfReaderEndpoint      string
	WorkerTimeoutInSeconds int64
	StaticContentEndpoint  string
}

// Run the lisener and request's router, activate rest server
func (s Server) Run(ctx context.Context) error {
	log.Printf("[INFO] activate rest server")

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		if httpServer != nil {
			if clsErr := httpServer.Close(); clsErr != nil {
				log.Printf("[ERROR] failed to close proxy http server, %v", clsErr)
			}
		}
	}()

	err := httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)

	if !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "server failed")
	}
	return nil
}

func (s Server) routes() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Throttle(1000), middleware.Timeout(60*time.Second))
	router.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))

	router.Route("/api/v1", func(r chi.Router) {
		r.Use(Logger(log.Default()))
		r.Get("/recipes", s.getRecepiesCtrl)
		r.Post("/recipes/sync", s.syncRecepiesCtrl)
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1") {
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, JSON{"error": "not found"})
			return
		}

		s.render(w, http.StatusNotFound, "404.tmpl.html", baseTmpl, "not found")
	})

	router.Group(func(r chi.Router) {
		r.Use(Logger(log.Default()))
		r.Use(middleware.StripSlashes)
		r.Get("/", s.indexCtrl)

		r.Post("/servings/cooked", s.cookedViewCtrl)

		r.Get("/recipes", s.recipesViewCtrl)
		r.Get("/recipes/more", s.moreRecipesViewCtrl)
		r.Post("/recipes/select", s.selectRecipeViewCtrl)

		r.Get("/pantry", s.pantryViewCtrl)
		r.Get("/pantry/add", s.ingredientFormViewCtrl)
		//r.Post("/pantry/add", s.createIngredientViewCtrl) todo move to pentry logic version 0.0.2
		r.Get("/pantry/edit", s.editIngredientViewCtrl)
	})

	// Serve static JavaScript and CSS files from the embedded content
	router.Get("/dist/*", func(w http.ResponseWriter, r *http.Request) {
		// Extract the requested file path after "/dist/"
		filePath := chi.URLParam(r, "*")

		// Read the file from the embedded content
		fileData, err := frontend.Assets.ReadFile("dist/" + filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("File not found: %s", filePath), http.StatusNotFound)
			return
		}

		// Determine the content type based on the file extension
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		w.Header().Set("Content-Type", contentType)

		// Write the file content to the response
		w.Write(fileData)
	})

	// Serve static data files from the data folder
	router.Get("/data/*", func(w http.ResponseWriter, r *http.Request) {
		// Extract the requested file path after "/data/"
		filePath := chi.URLParam(r, "*")

		filePath = "data/" + filePath
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("File not found: %s", filePath), http.StatusNotFound)
			return
		}

		// Determine the content type based on the file extension
		contentType := mime.TypeByExtension(filepath.Ext(filePath))
		w.Header().Set("Content-Type", contentType)

		// Write the file content to the response
		w.Write(fileData)
	})

	return router
}

func parseQueryParam(param string) (int, error) {
	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func renderBadRequest(w http.ResponseWriter, r *http.Request, message string, err error) {
	log.Printf("[ERROR] %s: %v", message, err)
	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, JSON{"error": err.Error(), "message": message})
}

func renderInternalServerError(w http.ResponseWriter, r *http.Request, message string, err error) {
	log.Printf("[ERROR] %s: %v", message, err)
	render.Status(r, http.StatusInternalServerError)
	render.JSON(w, r, JSON{"error": err.Error(), "message": message})
}
