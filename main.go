package main

import (
	"log"
	"net/http"
)

func main() {
	const port = "8080"

	serveMux := http.NewServeMux()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}
