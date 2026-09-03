package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/SlothEfficiency/chirpy/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Could not connect to database.")
		os.Exit(1)
	}

	apiConfig := apiConfig{
		db: database.New(db),
	}
	mux := http.NewServeMux()

	defaultHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiConfig.middleWareMetricInc(defaultHandler))
	mux.HandleFunc("GET /api/healthz/", customHandler)
	mux.HandleFunc("GET /admin/metrics/", apiConfig.getHitsHandler)
	mux.HandleFunc("POST /admin/reset/", apiConfig.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
