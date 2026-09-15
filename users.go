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
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
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
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

type upgradeWebhook struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
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
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
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
			IsChirpyRed:  user.IsChirpyRed,
			Token:        token,
			RefreshToken: refreshToken,
		}
		sendResponse(w, 200, safeUser)
		return
	}
	w.WriteHeader(401)
	w.Write([]byte("Incorrect email or password"))
}

func (cfg *apiConfig) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
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

	decoder := json.NewDecoder(r.Body)
	newEmailPassword := createUserRequest{}
	err = decoder.Decode(&newEmailPassword)
	if err != nil {
		sendError(w, "request could not be decoded.", 400, err)
		return
	}

	hashedPassword, err := auth.HashPassword(newEmailPassword.Password)
	if err != nil {
		sendError(w, "Password could not be hashed.", 500, err)
		return
	}

	params := database.UpdateEmailPasswordParams{
		ID:              userID,
		Email:           newEmailPassword.Email,
		HashedPasswords: hashedPassword,
	}
	newUser, err := cfg.db.UpdateEmailPassword(r.Context(), params)
	if err != nil {
		sendError(w, "Password could not be updated.", 500, err)
		return
	}
	userResponse := createUserResponse{
		ID:          newUser.ID,
		CreatedAt:   newUser.CreatedAt,
		UpdatedAt:   newUser.UpdatedAt,
		Email:       newUser.Email,
		IsChirpyRed: newUser.IsChirpyRed,
	}
	sendResponse(w, 200, userResponse)
}

func (cfg *apiConfig) upgradeUserHandler(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		sendError(w, "APIKey could not be extracted.", 401, err)
		return
	}
	if apiKey != cfg.polkaKey {
		sendError(w, "APIKey was wrong.", 401, err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	webhook := upgradeWebhook{}
	err = decoder.Decode(&webhook)
	if err != nil {
		sendError(w, "request could not be decoded.", 400, err)
		return
	}

	if webhook.Event != "user.upgraded" {
		sendResponse(w, 204, nil)
		return
	}

	_, err = cfg.db.UpgradeUserByID(r.Context(), webhook.Data.UserID)
	if err != nil {
		sendError(w, "User could not be upgraded", 404, err)
		return
	}
	sendResponse(w, 204, nil)
}
