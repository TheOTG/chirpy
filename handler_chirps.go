package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/TheOTG/chirpy/internal/auth"
	"github.com/TheOTG/chirpy/internal/database"
	"github.com/google/uuid"
)

func validateChirp(body string) (string, error) {
	if len(body) > 140 {
		return "", errors.New("chirp is too long")
	}

	replaceMap := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	splitStr := strings.Split(body, " ")
	for i, str := range splitStr {
		lowerCase := strings.ToLower(str)
		for k := range replaceMap {
			if strings.Contains(lowerCase, k) {
				splitStr[i] = "****"
				break
			}
		}
	}

	return strings.Join(splitStr, " "), nil
}

func (cfg *apiConfig) handlerListChirps(w http.ResponseWriter, r *http.Request) {
	authorID := uuid.NullUUID{Valid: false}
	q1 := r.URL.Query().Get("author_id")
	if q1 != "" {
		parsed, err := uuid.Parse(q1)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Error parsing id", err)
			return
		}
		authorID.UUID = parsed
		authorID.Valid = true
	}

	q2 := r.URL.Query().Get("sort")

	chirps, err := cfg.db.ListChirps(r.Context(), authorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps", err)
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserID    uuid.UUID `json:"user_id"`
	}

	respBody := []response{}

	for _, chirp := range chirps {
		respBody = append(respBody, response{
			ID:        chirp.ID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			UserID:    chirp.UserID,
		})
	}

	if q2 == "desc" {
		slices.Reverse(respBody)
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp id", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserID    uuid.UUID `json:"user_id"`
	}

	respBody := response{
		ID:        chirp.ID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   cleaned,
		UserID: id,
	}
	chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserID    uuid.UUID `json:"user_id"`
	}

	respBody := Chirp{
		ID:        chirp.ID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusCreated, respBody)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp id", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
