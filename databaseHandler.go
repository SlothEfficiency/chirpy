package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type createUserRequest struct {
	Email string `json:"email"`
}

type createUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userRequest := createUserRequest{}

	err := decoder.Decode(&userRequest)
	if err != nil {
		sendError(w, err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), userRequest.Email)
	if err != nil {
		sendError(w, err)
		return
	}

	userResponse := createUserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	res, err := json.Marshal(userResponse)
	if err != nil {
		sendError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(res)
}
