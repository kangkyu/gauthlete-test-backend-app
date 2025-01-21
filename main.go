package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", validateToken(homeHandler))

	log.Println("Starting backend server on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}

func validateToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No token provided", http.StatusUnauthorized)
			return
		}

		// Call auth server's introspection endpoint
		req, err := http.NewRequest("POST", "http://localhost:8080/introspect", nil)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		// Forward the same Authorization header
		req.Header.Set("Authorization", authHeader)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling introspection: %v", err)
			http.Error(w, "Failed to validate token", http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Introspection returned status: %d", resp.StatusCode)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		var introspectionResp struct {
			Usable  bool     `json:"usable"`
			Subject string   `json:"subject"`
			Scopes  []string `json:"scopes"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&introspectionResp); err != nil {
			log.Printf("Error decoding introspection response: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		if !introspectionResp.Usable {
			http.Error(w, "Token not active", http.StatusUnauthorized)
			return
		}

		log.Printf("Token validated successfully for subject: %s", introspectionResp.Subject)
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
