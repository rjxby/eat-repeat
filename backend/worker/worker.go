package worker

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rjxby/eat-repeat/backend/store"
)

// Error messages
var (
	ErrUpdateJobStatus = fmt.Errorf("failed to update job status")
)

// WorkerProc runs jobs
type WorkerProc struct {
	pdfReaderEndpoint      string
	workerTimeoutInSeconds int64
	engine                 Engine
}

// New creates WorkerProc
func New(pdfReaderEndpoint string, workerTimeoutInSeconds int64, engine Engine) *WorkerProc {
	return &WorkerProc{
		pdfReaderEndpoint:      pdfReaderEndpoint,
		workerTimeoutInSeconds: workerTimeoutInSeconds,
		engine:                 engine,
	}
}

// Engine defines an interface to save and load recipes
type Engine interface {
	SaveSyncJob(job *store.JobV1) (*store.JobV1, error)
	LoadRecipes(page, pageSize int, searchTerm string) (*store.Recipes, error)
	SaveRecipe(recipe *store.RecipeV1) (*store.RecipeV1, error)
}

type RecipeJSON struct {
	Title       string        `json:"title"`
	SubTitle    string        `json:"sub_title"`
	Description string        `json:"description"`
	CookTime    uint          `json:"cook_time"`
	Nutrition   NutritionJSON `json:"nutrition_info"`
}

type NutritionJSON struct {
	Calories uint `json:"calories_per_serving"`
	Carbs    uint `json:"net_carbs_per_serving"`
}

// RunSyncRecipes runs the synchronization of recipes
func (p *WorkerProc) RunSyncRecipes() (*store.JobV1, error) {
	job := store.JobV1{
		Status:    store.JobStatusPending,
		CreatedAt: time.Now().UTC(),
	}

	savedJob, err := p.engine.SaveSyncJob(&job)
	if err != nil {
		return nil, err
	}

	go runSyncRecipes(p, savedJob)

	return savedJob, nil
}

func runSyncRecipes(p *WorkerProc, job *store.JobV1) {
	job.Status = store.JobStatusInProgress
	if _, err := p.engine.SaveSyncJob(job); err != nil {
		log.Print(ErrUpdateJobStatus)
	}

	recipesToSync, err := p.engine.LoadRecipes(math.MaxInt64, math.MaxInt64, "")
	if err != nil {
		log.Printf("[ERROR] failed to load recipes: %v", err)
		updateJobStatus(p, job, store.JobStatusFailed)
		return
	}

	resultCh := make(chan *store.RecipeV1)
	errorCh := make(chan error)

	var wg sync.WaitGroup

	for _, recipe := range recipesToSync.Recipes {

		// give a break to server, it's not a cloud scale set
		time.Sleep(time.Duration(2000) * time.Millisecond)

		wg.Add(1)
		go mapRecipeDetails(recipe, p.pdfReaderEndpoint, resultCh, errorCh, &wg)
	}

	go func() {
		wg.Wait()

		close(resultCh)
		close(errorCh)

		updateJobStatus(p, job, store.JobStatusCompleted)
	}()

	timeout := time.After(time.Duration(p.workerTimeoutInSeconds) * time.Second)
	for i := 0; i < len(recipesToSync.Recipes); i++ {
		select {
		case result := <-resultCh:
			if result != nil {
				if _, err := p.engine.SaveRecipe(result); err != nil {
					log.Printf("[ERROR] failed to update recipe (ID: %d) %s", result.ID, err)
					updateJobStatus(p, job, store.JobStatusFailed)
				}
			}
		case err := <-errorCh:
			log.Printf("[ERROR] failed to process recipe %s", err)
		case <-timeout:
			log.Printf("[ERROR] timeout to process batch in %d seconds", p.workerTimeoutInSeconds)
			updateJobStatus(p, job, store.JobStatusFailed)
			return
		}
	}
}

func updateJobStatus(p *WorkerProc, jobToUpdate *store.JobV1, status store.JobStatus) {
	jobToUpdate.Status = status
	if _, err := p.engine.SaveSyncJob(jobToUpdate); err != nil {
		log.Print(ErrUpdateJobStatus)
	}
}

func mapRecipeDetails(recipe store.RecipeV1, pdfReaderEndpoint string, resultCh chan<- *store.RecipeV1, errorCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	if !recipe.PdfUrl.Valid {
		return
	}

	file, err := os.Open(recipe.PdfUrl.String)
	if err != nil {
		log.Printf("[ERROR] failed to open recipe pdf (ID: %d)", recipe.ID)
		errorCh <- err
		return
	}
	defer file.Close()

	body, contenType, err := createMultipartRequestBody(file, recipe.PdfUrl.String)
	if err != nil {
		log.Printf("[ERROR] failed to create request for recipe pdf (ID: %d)", recipe.ID)
		errorCh <- err
		return
	}

	responseBody, err := sendPostRequest(pdfReaderEndpoint, body, contenType)
	if err != nil {
		log.Printf("[ERROR] failed to send request for recipe pdf (ID: %d)", recipe.ID)
		errorCh <- err
		return
	}

	recipeDetails, err := processResponse(responseBody)
	if err != nil {
		log.Printf("[ERROR] failed to process response for recipe pdf (ID: %d)", recipe.ID)
		errorCh <- err
		return
	}

	updateRecipe := mapRecipe(&recipe, recipeDetails)
	resultCh <- updateRecipe
}

func mapRecipe(destination *store.RecipeV1, source *RecipeJSON) (result *store.RecipeV1) {
	if source.Title != "" {
		destination.Title = source.Title
	}

	destination.Description = source.Description
	destination.CookingTimeInMinutes = source.CookTime
	destination.UpdatedAt = sql.NullTime{
		Time:  time.Now().UTC(),
		Valid: true,
	}

	return destination
}

func createMultipartRequestBody(file *os.File, fileName string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileField, err := writer.CreateFormFile("file", filepath.Base(fileName))
	if err != nil {
		log.Print("[ERROR] failed to create form file")
		return nil, "", err
	}

	_, err = io.Copy(fileField, file)
	if err != nil {
		log.Print("[ERROR] failed to copy file content")
		return nil, "", err
	}

	contentType := writer.FormDataContentType()
	writer.Close()
	return body, contentType, nil
}

func sendPostRequest(url string, body *bytes.Buffer, contenType string) ([]byte, error) {
	response, err := http.Post(url, contenType, body)
	if err != nil {
		log.Printf("[ERROR] failed to send request: %v", err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Printf("[ERROR] unexpected status code: %d", response.StatusCode)
		return nil, fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("[ERROR] failed to read response body: %v", err)
		return nil, err
	}

	return responseBody, nil
}

func processResponse(responseBody []byte) (*RecipeJSON, error) {
	var recipeDetails RecipeJSON
	err := json.Unmarshal(responseBody, &recipeDetails)
	if err != nil {
		log.Print("[ERROR] failed to unmarshal JSON")
		return nil, err
	}

	return &recipeDetails, nil
}
