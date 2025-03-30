package route

import (
	"api.cap.iot/service"
	"encoding/json"
	"net/http"
)

// LoginRequest is the struct for the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SetupAuthRoutes initializes the authentication routes
func SetupAuthRoutes(mux *http.ServeMux, authService *service.AuthService) {
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, authService)
	})
}

// LoginHandler handles login requests and delegates logic to AuthService
func LoginHandler(w http.ResponseWriter, r *http.Request, authService *service.AuthService) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the login request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call AuthService to authenticate and generate JWT
	token, err := authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Respond with the JWT
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"jwtToken": token})
}
