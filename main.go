package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/TheOTG/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	secretKey      string
	platform       string
	polkaKey       string
}

func (cfg *apiConfig) getFileServerHits() int32 {
	return cfg.fileserverHits.Load()
}

func (cfg *apiConfig) resetHitCount() {
	cfg.fileserverHits.Swap(0)
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secretKey := os.Getenv("SECRET_KEY")
	polkaKey := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	serveMux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: serveMux,
	}

	apiCfg := &apiConfig{}
	apiCfg.db = dbQueries
	apiCfg.secretKey = secretKey
	apiCfg.platform = platform
	apiCfg.polkaKey = polkaKey

	fileServer := http.FileServer(http.Dir("."))
	appHandler := http.StripPrefix("/app/", fileServer)

	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(appHandler))

	serveMux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerPolkaWebhook)

	serveMux.HandleFunc("GET /api/healthz", apiCfg.handlerHealth)

	serveMux.HandleFunc("GET /api/chirps", apiCfg.handlerListChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)

	serveMux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	serveMux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)
	serveMux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	serveMux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	log.Println("Starting server...")
	log.Fatal(server.ListenAndServe())
}
