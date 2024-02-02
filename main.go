package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/rjxby/eat-repeat/backend/chef"
	"github.com/rjxby/eat-repeat/backend/pantry"
	"github.com/rjxby/eat-repeat/backend/scheduler"
	"github.com/rjxby/eat-repeat/backend/server"
	"github.com/rjxby/eat-repeat/backend/store"
	"github.com/rjxby/eat-repeat/backend/worker"
)

var revision string

func main() {
	log.Printf("eat-repeat %s\n", revision)

	appSettings := parseEnvironment()

	templateCache, err := server.NewTemplateCache()
	if err != nil {
		log.Printf("[ERROR] can't create template cache, %+v", err)
		os.Exit(1)
	}

	dataStore, err := getEngine(appSettings.RunMigration)
	if err != nil {
		log.Printf("[ERROR] can't create data store, %+v", err)
		os.Exit(1)
	}

	srv := server.Server{
		Chef:          chef.New(dataStore),
		Scheduler:     scheduler.New(),
		Pantry:        pantry.New(dataStore),
		Worker:        worker.New(appSettings.PdfReaderEndpoint, appSettings.WorkerTimeoutInSeconds, dataStore),
		Version:       revision,
		TemplateCache: templateCache,
	}

	if err := srv.Run(context.Background()); err != nil {
		log.Printf("[ERROR] failed, %+v", err)
	}
}

func getEngine(runMigration bool) (*store.Database, error) {

	database, err := store.NewDatabase()
	if err != nil {
		log.Fatalf("[ERROR] can't open db, %v", err)
		return nil, err
	}

	if runMigration {
		err = database.Migrate()
		if err != nil {
			log.Fatalf("[ERROR] can't migrate db, %v", err)
			return nil, err
		}
	}

	return database, nil
}

type Settings struct {
	RunMigration           bool
	PdfReaderEndpoint      string
	WorkerTimeoutInSeconds int64
}

func parseEnvironment() Settings {
	settings := Settings{
		RunMigration:           false,
		WorkerTimeoutInSeconds: 900, // 900s = 15 min
	}

	runMigrationStr := os.Getenv("RUN_MIGRATION")
	if runMigrationStr != "" {
		var err error
		settings.RunMigration, err = strconv.ParseBool(runMigrationStr)
		if err != nil {
			fmt.Println("Error parsing RUN_MIGRATION environment variable:", err)
		}
	}

	pdfReaderEndpointStr := os.Getenv("PDF_READER_ENDPOINT")
	if pdfReaderEndpointStr == "" {
		log.Fatal("PDF_READER_ENDPOINT environment variable is not set")
	} else {
		settings.PdfReaderEndpoint = pdfReaderEndpointStr
	}

	workerTimeoutInSecondsStr := os.Getenv("WORKER_TIMEOUT_IN_SECONDS")
	if workerTimeoutInSecondsStr != "" {
		var err error
		settings.WorkerTimeoutInSeconds, err = strconv.ParseInt(workerTimeoutInSecondsStr, 10, 64)
		if err != nil {
			fmt.Println("Error parsing WORKER_TIMEOUT_IN_SECONDS environment variable:", err)
		}
	}

	return settings
}
