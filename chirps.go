package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/SlothEfficiency/chirpy/internal/database"
	"github.com/google/uuid"
)

func customHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

type ChirpRequest struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	chirp := ChirpRequest{}
	err := decoder.Decode(&chirp)
	if err != nil {
		sendError(w, "request could not be decoded.", 400, err)
		return
	}

	err = validateChirp(chirp)
	if err != nil {
		sendError(w, "chirp was too long", 400, err)
		return
	}

	parameters := database.CreateChirpParams{
		Body:   replaceProfaneWords(chirp.Body),
		UserID: chirp.UserID,
	}
	chirpEntry, err := cfg.db.CreateChirp(r.Context(), parameters)
	if err != nil {
		sendError(w, "Chirp creation failed", 500, err)
		return
	}

	chirpResponse := ChirpResponse{
		ID:        chirpEntry.ID,
		CreatedAt: chirpEntry.CreatedAt,
		UpdatedAt: chirpEntry.UpdatedAt,
		Body:      chirpEntry.Body,
		UserID:    chirpEntry.UserID,
	}
	sendResponse(w, 201, chirpResponse)
}

func validateChirp(chirp ChirpRequest) error {
	if len(chirp.Body) > 140 {
		fmt.Println("chirp too long")
		return fmt.Errorf("Chirp is too long")
	}
	return nil
}

func replaceProfaneWords(input string) string {
	output := input
	profaneWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Fields(input)
	for _, word := range words {
		if profaneWords[strings.ToLower(word)] {
			output = strings.Replace(output, word, "****", 1)
		}
	}
	return output
}

func (cfg *apiConfig) getAllChirpsHandler(w http.ResponseWriter, r *http.Request) {
	allChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		sendError(w, "Couldn't get all chirps", 500, err)
		return
	}

	formatedChirps := []ChirpResponse{}
	for _, chirp := range allChirps {
		formatedChirps = append(formatedChirps, ChirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	sendResponse(w, 200, formatedChirps)
}
