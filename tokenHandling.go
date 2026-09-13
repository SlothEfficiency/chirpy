package main

import (
	"net/http"
	"time"

	"github.com/SlothEfficiency/chirpy/internal/auth"
)

type refreshResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		sendError(w, "Couldnt read token.", 401, err)
		return
	}

	userID, err := cfg.db.GetUserByRefreshToken(r.Context(), refreshToken)
	if err != nil {
		sendError(w, "Couldnt find active refresh token.", 401, err)
		return
	}

	token, err := auth.MakeJWT(userID, cfg.tokenSecret, 60*time.Minute)
	if err != nil {
		sendError(w, "Couldnt create token.", 500, err)
		return
	}
	response := refreshResponse{
		Token: token,
	}
	sendResponse(w, 200, response)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		sendError(w, "Couldnt read token.", 401, err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		sendError(w, "Couldnt revoke refresh token.", 401, err)
		return
	}

	sendResponse(w, 204, nil)
}
