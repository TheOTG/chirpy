package main

import (
	"net/http"
	"time"

	"github.com/TheOTG/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing token", err)
		return
	}

	refreshToken, err := cfg.db.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secretKey, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating access token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}

	respBody := response{
		Token: accessToken,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Missing token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error revoking token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
