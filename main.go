package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/rowinf/chirpy/internal"
)

type apiConfig struct {
	fileServerHits int
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	type errorBody struct {
		Error string `json:"error"`
	}

	errBody := errorBody{
		Error: message,
	}
	dat, err := json.Marshal(errBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func CensorString(s string) string {
	censorship := []byte("****")
	bytes := []byte(s)
	re := regexp.MustCompile("(?i)kerfuffle|sharbert|fornax")
	val := re.ReplaceAll(bytes, censorship)
	return string(val)
}

func createChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type returnVals struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	body := CensorString(string(params.Body))
	if err != nil {
		log.Printf("error decoding parameters %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if len(body) > 140 {
		respondWithError(w, http.StatusBadRequest, "chirp longer than 140 characters")
		return
	}
	db, err := internal.NewDB("./database/database.json")
	if err != nil {
		panic("error database")
	}
	if chirp, err := db.CreateChirp(body); err == nil {
		respondWithJSON(w, http.StatusOK, chirp)
	} else {
		respondWithError(w, 400, "unprocessable chirp")
	}
}

func main() {
	port := "8080"

	apiConfig := apiConfig{
		fileServerHits: 0,
	}
	r := chi.NewRouter()
	api := chi.NewRouter()
	admin := chi.NewRouter()
	// Create a new ServeMux
	handler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	r.Get("/app", handler)
	r.Get("/app/*", handler)
	api.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
	api.Post("/chirps", createChirp)
	admin.Get("/metrics", adminMetrics(&apiConfig))
	admin.Get("/metrics/", adminMetrics(&apiConfig))
	admin.Get("/reset", apiConfig.handlerReset)
	r.Mount("/api", api)
	r.Mount("/admin", admin)
	// Wrp the mux in a custom middleware for CORS
	corsMux := addCorsHeaders(r)

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

func adminMetrics(apiConfig *apiConfig) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile("./admin/index.html")
		if err != nil {
			http.Error(w, "missing index.html", http.StatusInternalServerError)
		}
		htmlContent := fmt.Sprintf(string(content), apiConfig.fileServerHits)
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	})
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.HandlerFunc {
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
