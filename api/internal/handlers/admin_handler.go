package handlers

import (
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type AdminHandler struct {
	authService     service.AuthService
	userService     service.UserService
	deviceService   service.DeviceService
	locationService service.LocationService
}

func NewAdminHandler(authService *service.DefaultAuthService, userService *service.DefaultUserService, deviceService *service.DefaultDeviceService, locationService service.LocationService) *AdminHandler {
	return &AdminHandler{
		authService:     authService,
		userService:     userService,
		deviceService:   deviceService,
		locationService: locationService,
	}
}

func (h *AdminHandler) GetDevicescomponentsLocations(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for devices, components, and locations", request.Method)

	if request.Method != http.MethodGet {
		log.Printf("Method %s not allowed...", request.Method)
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Call AuthService with pagination and search parameters
	devicesWithPagination, err := h.deviceService.GetDevicescomponentsLocations(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get data with pagination and search: %v", err)
		http.Error(writer, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(devicesWithPagination); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	}
}

func (h *AdminHandler) GetUsersLocations(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for users and locations", request.Method)

	if request.Method != http.MethodGet {
		log.Printf("Method %s not allowed...", request.Method)
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the user ID from the query parameters
	vars := mux.Vars(request)

	userIDStr, ok := vars["userID"]
	if !ok {
		log.Printf("Missing user ID in path")
		http.Error(writer, "Missing user ID in path", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		http.Error(writer, "Invalid user ID format", http.StatusBadRequest)
		return
	}
	// Call AuthService to get users and locations
	usersLocations, err := h.userService.GetUserSites(request.Context(), userID)
	if err != nil {
		log.Printf("Failed to get users and locations: %v", err)
		http.Error(writer, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(usersLocations); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	}
}

func (h *AdminHandler) GetLocationsDevicesUsers(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for locations, devices, and users", request.Method)

	if request.Method != http.MethodGet {
		log.Printf("Method %s not allowed...", request.Method)
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	// Call AuthService to get locations, devices, and users
	locDevUsers, err := h.locationService.GetLocationsDevicesUsers(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get locations, devices, and users: %v", err)
		http.Error(writer, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(locDevUsers); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
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
	// Retrieve the user claims structure from the context
	claims, ok := r.Context().Value(middleware.UserClaimsContextKey).(*utils.Claims)
	if !ok {
		// Log error if user claims are not found or are of the wrong type
		log.Println("User claims not found or invalid type in context, authentication middleware might be missing or misconfigured")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Access the User ID from the claims structure
	auth0UserID := claims.Sub
	log.Printf("Retrieved User ID from claims: %d", auth0UserID)

	if !ok {
		log.Printf("Invalid user ID type in context: %T, expected string", auth0UserID)
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

func (h *AdminHandler) AssignDeviceToLocation(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to assign device to location", request.Method)

	if request.Method != http.MethodPost {
		log.Printf("Method %s not allowed...", request.Method)
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var assignment models.AssignDeviceToLocationRequest
	if err := json.NewDecoder(request.Body).Decode(&assignment); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		utils.RespondWithError(writer, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if err := h.deviceService.SetDeviceToLocation(request.Context(), assignment.DeviceID, assignment.LocationID); err != nil {
		log.Printf("Failed to assign device to location: %v", err)
		utils.RespondWithError(writer, http.StatusInternalServerError, "Failed to assign device to location", nil)
		return
	}

	writer.WriteHeader(http.StatusNoContent) // Standard response for successful assignment with no body
}

func (h *AdminHandler) DeleteLocation(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to delete location", request.Method)

	if request.Method != http.MethodDelete {
		log.Printf("Method %s not allowed...", request.Method)
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	locationID, ok := vars["locationID"]
	if !ok {
		log.Printf("Missing location ID in path")
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing location ID in path", nil)
		return
	}

	log.Printf("Attempting to delete location with ID: %s", locationID)

	if err := h.locationService.DeleteLocation(request.Context(), locationID); err != nil {
		log.Printf("Failed to delete location '%s': %v", locationID, err)
		if apiErr, ok := err.(*models.APIError); ok {
			utils.RespondWithError(writer, apiErr.StatusCode, apiErr.Message, apiErr.Details)
			return
		}
		utils.RespondWithError(writer, http.StatusInternalServerError, "Failed to delete location", map[string]string{"error": err.Error(), "locationID": locationID})
		return
	}

	writer.WriteHeader(http.StatusNoContent) // Standard response for successful deletion with no body
}

func (h *AdminHandler) ModifyLocation(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to modify location", request.Method)

	if request.Method != http.MethodPut {
		log.Printf("Method %s not allowed...", request.Method)
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	locationIDStr, ok := vars["locationID"]
	if !ok {
		log.Printf("Missing location ID in path")
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing location ID in path", nil)
		return
	}

	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		log.Printf("Invalid location ID in path: %v", err)
		utils.RespondWithError(writer, http.StatusBadRequest, "Invalid location ID in path", nil)
		return
	}

	var location = &models.Location{}
	if err := json.NewDecoder(request.Body).Decode(location); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		utils.RespondWithError(writer, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}
	location.ID = &locationID // Set the ID from the URL path

	if err := h.locationService.ModifyLocation(request.Context(), *location); err != nil {
		log.Printf("Failed to modify location: %v", err)
		utils.RespondWithError(writer, http.StatusInternalServerError, "Failed to modify location", nil)
		return
	}

	writer.WriteHeader(http.StatusNoContent) // Standard response for successful modification with no body
}
