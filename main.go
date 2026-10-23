package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"

	"github.com/coderianx/avenor/internal/config"
	"github.com/coderianx/avenor/internal/handlers/providers"
	"github.com/coderianx/avenor/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	port := flag.String("port", ":8080", "port for run server")
	useLogger := flag.Bool("logger", false, "logger")

	flag.Parse()

	// init
	err := config.InitAvenor()

	if err != nil {
		log.Fatal(err)
	}

	// connect database
	err = store.ConnectDB(
		filepath.Join(config.GetAvenorPath(), "avenor.db"),
	)

	if err != nil {
		log.Fatal(err)
	}

	defer store.CloseDB()

	// create tables
	err = store.CreateTables()

	if err != nil {
		log.Fatal(err)
	}

	// chi router
	router := chi.NewRouter()

	if *useLogger {
		router.Use(middleware.Logger)
	}

	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/providers", providers.AddNewProvider)
		r.Get("/providers", providers.ListProviders)
	})

	// start server
	log.Printf("Server starting on the %s", *port)
	log.Fatal(http.ListenAndServe(*port, router))
}
