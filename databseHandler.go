package main

import (
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
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {

}
