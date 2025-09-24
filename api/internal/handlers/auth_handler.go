// internal/handlers/auth_handler.go
package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
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
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Decode the login request
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Log decoding error
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Call AuthService to authenticate and generate JWT
	token, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Invalid email or password", nil, http.StatusUnauthorized)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Respond with the JWT
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"jwtToken": token}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to send response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
	}
}

// RegisterHandler handles registration requests and delegates logic to AuthService
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Log incoming requests for debugging
	log.Printf("Received %s request for registration", r.Method)

	if r.Method != http.MethodPost {
		// Log method not allowed error
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Decode the registration request
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Log decoding error
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Call AuthService to register the user
	err := h.authService.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		// Log registration error
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to register user", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"message": "User registered successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log encoding error if the response fails to send
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to send response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
	}
}
