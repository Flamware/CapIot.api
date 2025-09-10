package handlers

import (
	"CapIot-api/internal/config"
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

// GetDevicescomponentsLocations handles the request to get all devices, components, and locations with pagination and search.
func (h *AdminHandler) GetDevicescomponentsLocations(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for devices, components, and locations", request.Method)

	if request.Method != http.MethodGet {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	pageStr := request.URL.Query().Get("page")
	limitStr := request.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	searchTerm := request.URL.Query().Get("search")

	devicesWithPagination, err := h.deviceService.GetDevicescomponentsLocations(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get data with pagination and search: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve data", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, devicesWithPagination)
}

// GetUsersLocations handles the request to get a specific user's sites.
func (h *AdminHandler) GetUsersLocations(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for users and locations", request.Method)

	if request.Method != http.MethodGet {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)

	userIDStr, ok := vars["userID"]
	if !ok {
		log.Printf("Missing user ID in path")
		apiErr := models.NewAPIError(models.ErrorCodeMissingParameter, "Missing user ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID format: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInvalidFormat, "Invalid user ID format", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	usersLocations, err := h.userService.GetUserSites(request.Context(), userID)
	if err != nil {
		log.Printf("Failed to get users and locations: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve data", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, usersLocations)
}

// GetLocationsDevicesUsers handles the request to get locations, devices, and users with pagination and search.
func (h *AdminHandler) GetLocationsDevicesUsers(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for locations, devices, and users", request.Method)

	if request.Method != http.MethodGet {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	pageStr := request.URL.Query().Get("page")
	limitStr := request.URL.Query().Get("limit")

	page := 1
	limit := 10

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	searchTerm := request.URL.Query().Get("search")

	locDevUsers, err := h.locationService.GetLocationsDevicesUsers(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get locations, devices, and users: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve data", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, locDevUsers)
}

// GetUserRoleHandler retrieves the roles of the authenticated user
// Note: This function belongs in a separate AuthHandler, but is updated here for consistency.
func (h *AuthHandler) GetUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for user roles", r.Method)

	if r.Method != http.MethodGet {
		log.Printf("Method %s not allowed for getting user roles", r.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	claims, ok := r.Context().Value(config.UserClaimsContextKey).(*utils.Claims)
	if !ok {
		log.Println("User claims not found or invalid type in context, authentication middleware might be missing or misconfigured")
		apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
		utils.RespondWithError(w, apiErr)
		return
	}

	auth0UserID := claims.Sub
	log.Printf("Retrieved User ID from claims: %d", auth0UserID)

	// Note: The original code had a redundant 'if !ok' check here. Removed for clarity.

	roles, err := h.authService.GetUserRoles(r.Context(), auth0UserID)
	if err != nil {
		log.Printf("Failed to get roles for user %s: %v", auth0UserID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve user roles", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	response := models.AuthResult{
		Auth0ID: auth0UserID,
		Role:    roles,
	}

	utils.RespondWithJSON(w, http.StatusOK, response)
}

// AssignDeviceToLocation handles the request to assign a device to a location.
func (h *AdminHandler) AssignDeviceToLocation(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to assign device to location", request.Method)

	if request.Method != http.MethodPost {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	var assignment models.AssignDeviceToLocationRequest
	if err := json.NewDecoder(request.Body).Decode(&assignment); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request payload", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	if err := h.deviceService.SetDeviceToLocation(request.Context(), assignment.DeviceID, assignment.LocationID); err != nil {
		log.Printf("Failed to assign device to location: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to assign device to location", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

// DeleteLocation handles the request to delete a location.
func (h *AdminHandler) DeleteLocation(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to delete location", request.Method)

	if request.Method != http.MethodDelete {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	locationID, ok := vars["locationID"]
	if !ok {
		log.Printf("Missing location ID in path")
		apiErr := models.NewAPIError(models.ErrorCodeMissingParameter, "Missing location ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	log.Printf("Attempting to delete location with ID: %s", locationID)

	if err := h.locationService.DeleteLocation(request.Context(), locationID); err != nil {
		log.Printf("Failed to delete location '%s': %v", locationID, err)
		// Assuming the service layer returns an APIError directly
		if apiErr, ok := err.(models.APIError); ok {
			utils.RespondWithError(writer, apiErr)
			return
		}
		// Fallback for other types of errors
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to delete location", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}

// ModifyLocation handles the request to modify a location.
func (h *AdminHandler) ModifyLocation(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to modify location", request.Method)

	if request.Method != http.MethodPut {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	locationIDStr, ok := vars["locationID"]
	if !ok {
		log.Printf("Missing location ID in path")
		apiErr := models.NewAPIError(models.ErrorCodeMissingParameter, "Missing location ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	locationID, err := strconv.Atoi(locationIDStr)
	if err != nil {
		log.Printf("Invalid location ID in path: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInvalidFormat, "Invalid location ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	var location = &models.Location{}
	if err := json.NewDecoder(request.Body).Decode(location); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request payload", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}
	location.ID = &locationID

	if err := h.locationService.ModifyLocation(request.Context(), *location); err != nil {
		log.Printf("Failed to modify location: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to modify location", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
