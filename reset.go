package main

import (
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}

	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error deleting users", err)
		return
	}
	cfg.resetHitCount()
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
