package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/rowinf/chirpy/internal"
)

type apiConfig struct {
	fileServerHits int
	jwtSecret      []byte
}

type UserParams struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

type MyCustomClaims struct {
	jwt.RegisteredClaims
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
	db, err := internal.NewDB("./database.json")
	if err != nil {
		panic("error database")
	}
	if chirp, err := db.CreateChirp(body); err == nil {
		respondWithJSON(w, http.StatusCreated, chirp)
	} else {
		respondWithError(w, 400, "unprocessable chirp")
	}
}

func getChirps(w http.ResponseWriter, r *http.Request) {
	db, err := internal.NewDB("./database.json")
	if err != nil {
		panic("error database")
	}

	if chirps, err := db.GetChirps(); err == nil {
		respondWithJSON(w, http.StatusOK, chirps)
	} else {
		respondWithError(w, 400, "unprocessable chirp")
	}
}

func getChirp(w http.ResponseWriter, r *http.Request) {
	db, err := internal.NewDB("./database.json")
	if err != nil {
		panic("error database")
	}
	chirpId, parseErr := strconv.Atoi(r.PathValue("chirpID"))
	if parseErr != nil {
		respondWithError(w, 404, "not found")
	} else if chirp, err := db.GetChirp(chirpId); err == nil {
		respondWithJSON(w, http.StatusOK, chirp)
	} else {
		respondWithError(w, 404, "not found")
	}
}

func GetTokenFromAuthorizationHeader(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("unauthorized")
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix)), nil
}

func updateUser(w http.ResponseWriter, r *http.Request, ctx *apiConfig) {
	decoder := json.NewDecoder(r.Body)
	params := UserParams{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("error decoding parameters %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	claims := MyCustomClaims{}
	authorization := r.Header.Get("Authorization")
	headerToken, herr := GetTokenFromAuthorizationHeader(authorization)
	if herr != nil {
		respondWithError(w, http.StatusUnauthorized, herr.Error())
		return
	}
	token, err := jwt.ParseWithClaims(headerToken, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("theres a problem with the signing method")
		}
		return ctx.jwtSecret, nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
	} else if db, dberr := internal.NewDB("./database.json"); dberr == nil {
		if token.Valid {
			userId, converr := strconv.Atoi(claims.Subject)
			if converr == nil {
				user, err := db.UpdateUser(userId, params.Email, []byte(params.Password))
				if err != nil {
					respondWithError(w, http.StatusBadRequest, "unprocessable user")
				} else {
					respondWithJSON(w, http.StatusOK, user)
				}
			} else {
				respondWithError(w, http.StatusBadRequest, "bad subject")
			}
		} else {
			respondWithError(w, http.StatusUnauthorized, "invalid token")
		}
	} else {
		respondWithError(w, http.StatusBadRequest, dberr.Error())
	}
}

func createUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	params := UserParams{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("error decoding parameters %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	db, _ := internal.NewDB("./database.json")
	user, err := db.CreateUser(params.Email, []byte(params.Password))

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "unprocessable user")
	} else {
		respondWithJSON(w, http.StatusCreated, user)
	}
}

func userLogin(w http.ResponseWriter, r *http.Request, ctx *apiConfig) {
	decoder := json.NewDecoder(r.Body)
	params := UserParams{}
	if err := decoder.Decode(&params); err != nil {
		log.Printf("error decoding parameters %s", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	tokenExpiration := time.Now().Add(24 * time.Hour)
	if params.ExpiresInSeconds > 0 {
		tokenExpiration = time.Now().Add(time.Duration(params.ExpiresInSeconds))
	}
	db, _ := internal.NewDB("./database.json")
	user, err := db.UserLogin(params.Email, []byte(params.Password))
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(tokenExpiration),
		Subject:   fmt.Sprint(user.Id),
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, serr := token.SignedString(ctx.jwtSecret)
	payload := struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}{
		Id:    user.Id,
		Email: user.Email,
		Token: ss,
	}
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
	} else if serr != nil {
		respondWithError(w, http.StatusInternalServerError, serr.Error())
	} else {
		respondWithJSON(w, http.StatusOK, payload)
	}
}

var dbg bool

func init() {
	flag.BoolVar(&dbg, "debug", false, "Enable debug mode")
}

func main() {
	godotenv.Load()
	flag.Parse()
	port := "8080"
	if dbg {
		os.Remove("./database.json")
	}

	apiConfig := apiConfig{
		fileServerHits: 0,
		jwtSecret:      []byte(os.Getenv("JWT_SECRET")),
	}
	r := http.NewServeMux()
	admin := http.NewServeMux()
	// Create a new ServeMux
	handler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	r.HandleFunc("/app", handler)
	r.HandleFunc("/app/*", handler)
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	})
	admin.HandleFunc("/metrics", adminMetrics(&apiConfig))
	admin.HandleFunc("/metrics/", adminMetrics(&apiConfig))
	admin.HandleFunc("/reset", apiConfig.handlerReset)
	r.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		userLogin(w, r, &apiConfig)
	})
	r.HandleFunc("POST /api/users", createUser)
	r.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		updateUser(w, r, &apiConfig)
	})
	r.HandleFunc("POST /api/chirps", createChirp)
	r.HandleFunc("GET /api/chirps", getChirps)
	r.HandleFunc("GET /api/chirps/{chirpID}", getChirp)
	r.Handle("/admin/", http.StripPrefix("/app", admin))
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
