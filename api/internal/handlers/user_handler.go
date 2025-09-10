// internal/handlers/user_handler.go
package handlers

import (
	"CapIot-api/internal/config"
	"CapIot-api/internal/service"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService *service.DefaultUserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		log.Println("GetCurrentUser: Invalid user claims type in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("GetCurrentUser: Retrieved claims from context: %+v", userClaims)

	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		log.Println("GetCurrentUser: Invalid user ID type in claims")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	userID := int(userIDFloat) // Convert float64 to int
	log.Printf("GetCurrentUser: Extracted User ID from claims: %d", userID)

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("GetCurrentUser: Failed to get user with ID %d: %v", userID, err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("GetCurrentUser: Failed to encode user to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("GetCurrentUser: Successfully returned current user")
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user claims", http.StatusInternalServerError)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	userID := int(userIDFloat)

	var requestBody struct {
		Name *string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	user.Name = requestBody.Name
	updatedUser, err := h.userService.UpdateUser(r.Context(), *user)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// AsignUser handles the request to assign a user to a location.
func (h *UserHandler) AsignUser(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL parameters
	vars := mux.Vars(r)
	userIDStr := vars["userID"]

	// Extract the location ID from the request body
	var requestBody struct {
		LocationID int `json:"locationID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert userID to int
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user ID and location ID are valid (greater than zero)
	if userIDInt <= 0 || requestBody.LocationID <= 0 {
		http.Error(w, "Invalid user ID or location ID", http.StatusBadRequest)
		return
	}

	// Call the service to assign the user to the location using the integer IDs
	err = h.userService.AsignUser(r.Context(), userIDInt, requestBody.LocationID)
	if err != nil {
		log.Printf("Failed to assign user %d to location %d: %v", userIDInt, requestBody.LocationID, err)
		http.Error(w, "Failed to assign user to location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // No content response
}

// GetUserLocations handles the request to retrieve all locations for a user.
func (h *UserHandler) GetUserLocations(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the JWT claims
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user claims", http.StatusInternalServerError)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	userID := int(userIDFloat)

	// Call the service to get user locations
	locations, err := h.userService.GetUserLocations(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get user locations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		log.Printf("GetUserLocations: Failed to encode locations to JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Println("GetUserLocations: Successfully returned user locations")
}

// UpdateUserAndLocation Function to update a user
func (h *UserHandler) UpdateUserAndLocation(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL parameters
	vars := mux.Vars(r)
	userIDStr := vars["userID"]

	// Extract the location ID from the request body
	var requestBody struct {
		SitesIDs []int  `json:"sites"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert userID to int
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user ID and location ID are valid (greater than zero)
	if userIDInt <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Call the service to assign the user to the location using the integer IDs
	err = h.userService.UpdateUserSites(r.Context(), userIDInt, requestBody.SitesIDs, requestBody.Name)
	if err != nil {
		log.Printf("Failed to update user %d to location %d: %v", userIDInt, requestBody.SitesIDs, err)
		http.Error(w, "Failed to update user to location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // No content response
}

func (h *UserHandler) GetUserSites(writer http.ResponseWriter, request *http.Request) {
	// Extract the user ID from the URL parameters
	vars := mux.Vars(request)
	userIDStr := vars["userID"]

	// Convert userID to int
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(writer, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user ID is valid (greater than zero)
	if userIDInt <= 0 {
		http.Error(writer, "Invalid user ID", http.StatusBadRequest)
		return
	}

	sites, err := h.userService.GetUserSites(request.Context(), userIDInt)
	if err != nil {
		http.Error(writer, "Failed to get user sites", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(sites); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) GetUsers(writer http.ResponseWriter, request *http.Request) {
	// Get pagination parameters from query string
	pageStr := request.URL.Query().Get("page")
	limitStr := request.URL.Query().Get("limit")

	page := 1
	limit := 10 // Default page size

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 { // Add a reasonable max limit
			limit = l
		}
	}

	// Get the optional search parameter
	searchTerm := request.URL.Query().Get("search")

	// Call AuthService to get all users
	users, err := h.userService.GetUsers(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get all users: %v", err)
		http.Error(writer, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(users); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) GetMySites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract the user ID from the JWT claims
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user claims", http.StatusInternalServerError)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	userID := int(userIDFloat)

	sites, err := h.userService.GetUserSites(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error getting sites", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sites); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// getNotifications retrieves notifications for the current user with Lazy Loading
func (h *UserHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the JWT claims
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid user claims", http.StatusInternalServerError)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}
	userID := int(userIDFloat)

	// Get pagination parameters from query string
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 10 // Default page size

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 { // Add a reasonable max limit
			limit = l
		}
	}

	notifications, err := h.userService.GetNotifications(r.Context(), userID, page, limit)
	if err != nil {
		http.Error(w, "Error getting notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
