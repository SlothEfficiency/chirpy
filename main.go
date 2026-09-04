package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/SlothEfficiency/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (cfg *apiConfig) middleWareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

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

	mux.HandleFunc("GET /admin/metrics", apiConfig.getHitsHandler)
	mux.HandleFunc("POST /admin/reset", apiConfig.resetHandler)

	mux.HandleFunc("GET /api/healthz", customHandler)

	mux.HandleFunc("GET /api/chirps", apiConfig.getAllChirpsHandler)
	mux.HandleFunc("POST /api/chirps", apiConfig.chirpHandler)

	mux.HandleFunc("POST /api/users", apiConfig.createUserHandler)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
