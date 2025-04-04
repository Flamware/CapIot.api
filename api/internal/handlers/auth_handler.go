// internal/handlers/auth_handler.go
package handlers

import (
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/models"
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
	token, err := h.authService.Login(req.Email, req.Password)
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
	err := h.authService.Register(req.Email, req.Password)
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

// GetUserRoleHandler retrieves the roles of the authenticated user
func (h *AuthHandler) GetUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	// Log incoming request
	log.Printf("Received %s request for user roles", r.Method)

	if r.Method != http.MethodGet {
		// Log method not allowed error
		log.Printf("Method %s not allowed for getting user roles", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from the context (set by JWTAuthMiddleware)
	userID := r.Context().Value(middleware.UserContextKey)
	if userID == nil {
		// Log error if user ID is not found in context
		log.Println("User ID not found in context, authentication middleware might be missing")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	auth0UserID, ok := userID.(string)
	if !ok {
		log.Printf("Invalid user ID type in context: %T, expected string", userID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Call AuthService to get user roles
	roles, err := h.authService.GetUserRoles(r.Context(), auth0UserID)
	if err != nil {
		// Log error from AuthService
		log.Printf("Failed to get roles for user %s: %v", auth0UserID, err)
		http.Error(w, "Failed to retrieve user roles", http.StatusInternalServerError)
		return
	}

	// Respond with the user ID and their roles
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := models.AuthResult{
		Auth0ID: auth0UserID,
		Role:    roles,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log encoding error
		log.Printf("Failed to encode user role response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
	}
}
