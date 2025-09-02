// internal/handlers/auth_handler.go
package handlers

import (
	"CapIot-api/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

// LoginRequest is the struct for the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the struct for the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// LoginHandler handles login requests and delegates logic to AuthService
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Log incoming requests for debugging
	log.Printf("Received %s request for login", r.Method)

	if r.Method != http.MethodPost {
		// Log method not allowed error
		log.Printf("Method %s not allowed for login", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the login request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Log decoding error
		log.Printf("Failed to decode login request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call AuthService to authenticate and generate JWT
	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		// Log authentication error
		log.Printf("Authentication failed for email %s: %v", req.Email, err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Respond with the JWT
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"jwtToken": token}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log encoding error if the response fails to send
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}

// RegisterHandler handles registration requests and delegates logic to AuthService
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Log incoming requests for debugging
	log.Printf("Received %s request for registration", r.Method)

	if r.Method != http.MethodPost {
		// Log method not allowed error
		log.Printf("Method %s not allowed for registration", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode the registration request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Log decoding error
		log.Printf("Failed to decode registration request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call AuthService to register the user
	err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		// Log registration error
		log.Printf("Registration failed for email %s: %v", req.Email, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"message": "User registered successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log encoding error if the response fails to send
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}
