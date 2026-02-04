package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/cors"
)

// Configuration
var jwtSecret = []byte("your-secret-key-change-in-production")

// Models
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type LoginResponse struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type CarListing struct {
	ID          int    `json:"id"`
	Make        string `json:"make"`
	Model       string `json:"model"`
	Year        int    `json:"year"`
	Price       string `json:"price"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Seller      string `json:"seller"`
}

// In-memory store
var (
	carStore = make(map[int]CarListing)
	nextID   = 1
	storeMu  sync.Mutex
)

// JWT Helper Functions
func generateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// Response Helpers
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Middleware
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}

func methodMiddleware(allowedMethod string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		next(w, r)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := validateJWT(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), "claims", claims)
		next(w, r.WithContext(ctx))
	}
}

// Handlers
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if user.Username == "seller" && user.Password == "carmarket123" {
		token, _ := generateJWT(user.Username)
		respondWithJSON(w, http.StatusOK, LoginResponse{Token: token, Message: "Login successful"})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
	}
}

func getCars(w http.ResponseWriter, r *http.Request) {
	storeMu.Lock()
	defer storeMu.Unlock()

	var listings []CarListing
	for _, car := range carStore {
		listings = append(listings, car)
	}
	respondWithJSON(w, http.StatusOK, listings)
}

func addCar(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	var car CarListing
	err := json.NewDecoder(r.Body).Decode(&car)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	car.ID = nextID
	car.Seller = claims.Username
	carStore[car.ID] = car
	nextID++

	respondWithJSON(w, http.StatusCreated, car)
}

// Page Handlers
func landingHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "main1.html"))
}

func loginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "login.html"))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "dashboard.html"))
}

func listingsPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("static", "listings.html"))
}

func main() {
	// CORS configuration
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/", loggingMiddleware(landingHandler))
	mux.HandleFunc("/login", loggingMiddleware(loginPageHandler))
	mux.HandleFunc("/api/login", loggingMiddleware(methodMiddleware("POST", loginHandler)))

	// Protected page routes
	mux.HandleFunc("/dashboard", loggingMiddleware(dashboardHandler))
	mux.HandleFunc("/listings", loggingMiddleware(listingsPageHandler))

	// Protected API routes
	mux.HandleFunc("/api/cars", loggingMiddleware(authMiddleware(methodMiddleware("GET", getCars))))
	mux.HandleFunc("/api/cars/add", loggingMiddleware(authMiddleware(methodMiddleware("POST", addCar))))

	// Add initial demo data
	storeMu.Lock()
	carStore[1] = CarListing{
		ID:          1,
		Make:        "McLaren",
		Model:       "765LT",
		Year:        2021,
		Price:       "$358,000",
		Description: "Extreme performance",
		ImageURL:    "https://example.com/image.jpg",
		Seller:      "demo",
	}
	nextID = 2
	storeMu.Unlock()

	// Server configuration
	port := "5001"
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      c.Handler(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server running on http://localhost:%s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
