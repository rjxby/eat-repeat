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
	RunMigration bool
}

func parseEnvironment() Settings {
	settings := Settings{
		RunMigration: false,
	}

	runMigrationStr := os.Getenv("RUN_MIGRATION")
	if runMigrationStr != "" {
		var err error
		settings.RunMigration, err = strconv.ParseBool(runMigrationStr)
		if err != nil {
			fmt.Println("Error parsing RUN_MIGRATION environment variable:", err)
		}
	}

	return settings
}
