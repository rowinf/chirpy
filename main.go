package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func main() {
	port := "8080"

	apiConfig := apiConfig{
		fileServerHits: 0,
	}
	// Create a new ServeMux
	mux := http.NewServeMux()
	handler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	mux.Handle("/app/", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
	mux.HandleFunc("/metrics/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hits: %d", apiConfig.fileServerHits)
	})
	mux.HandleFunc("/reset/", apiConfig.handlerReset)
	// Wrp the mux in a custom middleware for CORS
	corsMux := addCorsHeaders(mux)

	// Create a new HTTP server with the corsMux as the handler
	server := &http.Server{
		Addr:    ":" + port, // Set the desired port
		Handler: corsMux,
	}

	// Start the server
	log.Printf("Serving files from %s on port: %s\n", ".", port)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits = cfg.fileServerHits + 1
		next.ServeHTTP(w, r)
	})
}

// addCorsHeaders is a middleware function that adds CORS headers to the response.
func addCorsHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// If it's a preflight request, respond with 200 OK
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}
