package main

import (
	"database/sql"
	"fmt"
	"log"
	"main/internal/database"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	const port = "8080"

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	secret := os.Getenv("SECRET")
	if platform == "" {
		log.Fatal("SECRET must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Could not get connection to database: %s\n", err)
	}

	dbQueries := database.New(db)

	serveMux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
	}
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))

	serveMux.HandleFunc("GET /api/healthz", handleReadiness)
	serveMux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	serveMux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)

	serveMux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
