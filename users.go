package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/SlothEfficiency/chirpy/internal/auth"
	"github.com/SlothEfficiency/chirpy/internal/database"
	"github.com/google/uuid"
)

type createUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createUserResponse struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userRequest := createUserRequest{}

	err := decoder.Decode(&userRequest)
	if err != nil {
		sendError(w, "request could not be decoded.", 400, err)
		return
	}

	hashedPassword, err := auth.HashPassword(userRequest.Password)
	if err != nil {
		sendError(w, "Password could not be hashed.", 400, err)
		return
	}

	userRequestParams := database.CreateUserParams{
		Email:           userRequest.Email,
		HashedPasswords: hashedPassword,
	}
	user, err := cfg.db.CreateUser(r.Context(), userRequestParams)
	if err != nil {
		sendError(w, "User creation failed", 500, err)
		return
	}

	userResponse := createUserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	sendResponse(w, 201, userResponse)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	loginRequest := loginRequest{}

	err := decoder.Decode(&loginRequest)
	if err != nil {
		sendError(w, "request could not be decoded.", 400, err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), loginRequest.Email)
	if err != nil {
		sendError(w, "User was not found.", 401, err)
		return
	}

	authorized, err := auth.CheckPasswordHash(loginRequest.Password, user.HashedPasswords)
	if err != nil {
		sendError(w, "Passwords could not be compared.", 400, err)
		return
	}

	if authorized {
		token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, 3600*time.Second)
		if err != nil {
			sendError(w, "Couldnt generate token.", 500, err)
			return
		}

		refreshToken := auth.MakeRefreshToken()
		params := database.CreateRefreshTokenParams{
			Token:  refreshToken,
			UserID: user.ID,
		}
		refreshToken, err = cfg.db.CreateRefreshToken(r.Context(), params)
		if err != nil {
			sendError(w, "Couldnt save refresh token.", 500, err)
			return
		}

		safeUser := loginResponse{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			Token:        token,
			RefreshToken: refreshToken,
		}
		sendResponse(w, 200, safeUser)
		return
	}
	w.WriteHeader(401)
	w.Write([]byte("Incorrect email or password"))
}
