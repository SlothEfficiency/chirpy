package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/SlothEfficiency/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load(".env")
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
	mux.HandleFunc("POST /api/users", apiConfig.createUserHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}

func sendError(w http.ResponseWriter, err error) {
	errorResponse := ErrorResponse{
		Error: err.Error(),
	}
	byteResponse, err := json.Marshal(errorResponse)
	if err != nil {
		fmt.Println("Error couldnt be processed.")
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	w.Write(byteResponse)
}
