package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	const port = "8080"

	serveMux := http.NewServeMux()
	fileHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(fileHandler))
	serveMux.HandleFunc("GET /api/healthz", handleReadiness)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handleReadiness(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleMetrics(writer http.ResponseWriter, request *http.Request) {
	msg := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte(msg))
}

func (cfg *apiConfig) handleReset(writer http.ResponseWriter, request *http.Request) {
	cfg.fileServerHits.Store(0)
	writer.WriteHeader(200)
}
