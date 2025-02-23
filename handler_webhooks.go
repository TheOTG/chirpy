package main

import (
	"encoding/json"
	"net/http"

	"github.com/TheOTG/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error parsing id", err)
		return
	}

	_, err = cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	err = cfg.db.UpgradeUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error upgrading user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
