package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TheOTG/chirpy/internal/auth"
	"github.com/TheOTG/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error decoding parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting user", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secretKey, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating access token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token", err)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	}

	err = cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating refresh token in db", err)
		return
	}

	type response struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	respBody := response{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}
