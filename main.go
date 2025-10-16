package main

import (
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
	serveMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleReset)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
