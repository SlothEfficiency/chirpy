package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/SlothEfficiency/chirpy/internal/auth"
	"github.com/SlothEfficiency/chirpy/internal/database"
	"github.com/google/uuid"
)

func customHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

type ChirpRequest struct {
	Body string `json:"body"`
}

type ChirpResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) chirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		sendError(w, "Couldnt read token.", 400, err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		sendError(w, "Couldnt validate token", 401, err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	chirp := ChirpRequest{}
	err = decoder.Decode(&chirp)
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
		UserID: userID,
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
	authorID := r.URL.Query().Get("author_id")
	allChirps := []database.Chirp{}
	var err error
	if authorID != "" {
		authorUUID, err := uuid.Parse(authorID)
		if err != nil {
			sendError(w, "Author wrong format.", 404, err)
		}
		allChirps, err = cfg.db.GetAllChirpsFromAuthor(r.Context(), authorUUID)
	} else {
		allChirps, err = cfg.db.GetAllChirps(r.Context())
	}

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

	sort := r.URL.Query().Get("sort")
	if sort == "desc" {
		slices.SortFunc(formatedChirps, func(a, b ChirpResponse) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	}
	sendResponse(w, 200, formatedChirps)
}

func (cfg *apiConfig) getChirpByID(w http.ResponseWriter, r *http.Request) {
	pathValue, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, "Not valid uuid fromat", 400, err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), pathValue)
	if err != nil {
		sendError(w, "chirp not found", 404, err)
		return
	}

	formatedChirp := ChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	sendResponse(w, 200, formatedChirp)
}

func (cfg *apiConfig) deleteChirpHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		sendError(w, "Couldnt read token.", 401, err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.tokenSecret)
	if err != nil {
		sendError(w, "Couldnt validate token", 401, err)
		return
	}
	pathValue, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		sendError(w, "Not valid uuid fromat", 400, err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(r.Context(), pathValue)
	if err != nil {
		sendError(w, "chirp not found", 404, err)
		return
	}

	if chirp.UserID.String() != userID.String() {
		sendError(w, "Wrong user tried to delete a chirp", 403, fmt.Errorf("Wrong user tried to delete a chirp."))
		return
	}
	err = cfg.db.DeleteChirpbyID(r.Context(), pathValue)
	if err != nil {
		sendError(w, "chirp not deleted", 404, err)
		return
	}
	sendResponse(w, 204, nil)
}
