package handlers

import (
	"api.cap.iot/models"
	"api.cap.iot/service"
	"encoding/json"
	"fmt"
	"net/http"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginHandler handles the login request
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Step 1: Perform login and generate JWT
	customJwt, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Step 2: Return the JWT token to the frontend
	json.NewEncoder(w).Encode(map[string]string{"jwtToken": customJwt})
}
