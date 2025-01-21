package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/kangkyu/gauthlete"
)

var authleteClient *gauthlete.ServiceClient

func main() {
	// Initialize Authlete client
	authleteClient = gauthlete.NewServiceClient()

	mux := http.NewServeMux()

	// Protected route with authentication middleware
	mux.HandleFunc("/", validateToken(homeHandler))

	log.Println("Starting backend server on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

func validateToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No token provided", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token using Authlete
		resp, err := authleteClient.TokenIntrospect(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !resp.Usable {
			http.Error(w, "Token not usable", http.StatusUnauthorized)
			return
		}

		// Add user info to request context if needed
		// ctx := context.WithValue(r.Context(), "subject", resp.Subject)
		// next(w, r.WithContext(ctx))

		next(w, r)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"message": "Welcome to Protected Backend API Server",
		"status":  "authenticated",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
