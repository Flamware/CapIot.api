// internal/handlers/user_handler.go
package handlers

import (
	"CapIot-api/internal/config"
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
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
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Printf("GetCurrentUser: Retrieved claims from context: %+v", userClaims)

	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID in claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat) // Convert float64 to int
	log.Printf("GetCurrentUser: Extracted User ID from claims: %d", userID)

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get user", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error encoding response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Println("GetCurrentUser: Successfully returned current user")
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID in claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	var requestBody struct {
		Name *string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get user", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	user.Name = requestBody.Name
	updatedUser, err := h.userService.UpdateUser(r.Context(), *user)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to update user", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error encoding response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
}

// AsignUser godoc
// @Summary      Assign user to location
// @Description  Assigns a user to a specific location
// @Tags         users
// @Accept       json
// @Param        userID path int true "User ID"
// @Param        assignment body object{locationID=int} true "Assignment details"
// @Success      204  "No Content"
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/assign-user/{userID} [post]
func (h *UserHandler) AsignUser(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL parameters
	vars := mux.Vars(r)
	userIDStr := vars["userID"]

	// Extract the location ID from the request body
	var requestBody struct {
		locationID int `json:"locationID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Convert userID to int
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid user ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Check if the user ID and location ID are valid (greater than zero)
	if userIDInt <= 0 || requestBody.locationID <= 0 {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid user ID or location ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Call the service to assign the user to the location using the integer IDs
	err = h.userService.AsignUser(r.Context(), userIDInt, requestBody.locationID)
	if err != nil {
		log.Printf("Failed to assign user %d to location %d: %v", userIDInt, requestBody.locationID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to assign user to location", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent) // No content response
}

// GetUserLocations godoc
// @Summary      Get user locations
// @Description  Retrieves all locations assigned to the current user
// @Tags         users
// @Produce      json
// @Success      200  {array}   models.Location
// @Failure      500  {object}  models.APIError
// @Router       /api/users/me/locations [get]
func (h *UserHandler) GetUserLocations(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the JWT claims
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	// Call the service to get user locations
	locations, err := h.userService.GetUserLocations(r.Context(), userID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get user locations", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error encoding response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Println("GetUserLocations: Successfully returned user locations")
}

// UpdateUserAndLocation godoc
// @Summary      Update user and location
// @Description  Updates user details and assigned sites
// @Tags         users
// @Accept       json
// @Param        userID path int true "User ID"
// @Param        update body object{sites=[]int,name=string} true "Update details"
// @Success      204  "No Content"
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /admin/users/{userID} [put]
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
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Convert userID to int
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid user ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Check if the user ID and location ID are valid (greater than zero)
	if userIDInt <= 0 {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid user ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Call the service to assign the user to the location using the integer IDs
	err = h.userService.UpdateUserSites(r.Context(), userIDInt, requestBody.SitesIDs, requestBody.Name)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to update user and locations", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent) // No content response
}

// GetUserSites godoc
// @Summary      Get user sites
// @Description  Retrieves sites for a specific user
// @Tags         users
// @Produce      json
// @Param        userID path int true "User ID"
// @Success      200  {array}   models.Site
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /admin/users/{userID}/sites [get]
func (h *UserHandler) GetUserSites(writer http.ResponseWriter, request *http.Request) {
	// Extract the user ID from the URL parameters
	vars := mux.Vars(request)
	userIDStr := vars["userID"]

	// Convert userID to int
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid user ID", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Check if the user ID is valid (greater than zero)
	if userIDInt <= 0 {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid user ID", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	sites, err := h.userService.GetUserSites(request.Context(), userIDInt)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get user sites", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(sites); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error encoding response", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
}

// GetUsers godoc
// @Summary      Get all users
// @Description  Retrieves a paginated list of all users
// @Tags         users
// @Produce      json
// @Param        page query int false "Page number"
// @Param        limit query int false "Items per page"
// @Param        search query string false "Search term"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  models.APIError
// @Router       /admin/users [get]
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
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get users", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(users); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error encoding response", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
}

// GetMySites godoc
// @Summary      Get my sites
// @Description  Retrieves sites assigned to the current user
// @Tags         users
// @Produce      json
// @Success      200  {array}   models.Site
// @Failure      405  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/sites/me [get]
func (h *UserHandler) GetMySites(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}
	// Extract the user ID from the JWT claims
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	sites, err := h.userService.GetUserSites(r.Context(), userID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get user sites", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sites); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error encoding response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
}

// ChangeUsername godoc
// @Summary      Change username
// @Description  Updates the username of the current user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body body object{name=string} true "New username"
// @Success      200  {object}  models.User
// @Failure      400  {object}  models.APIError
// @Failure      500  {object}  models.APIError
// @Router       /api/users/me [put]
func (h *UserHandler) ChangeUsername(w http.ResponseWriter, r *http.Request) {
	// Extract the user ID from the URL parameters
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiError := models.APIError{
			Code:    models.ErrorCodeInternalServerError,
			Message: "Invalid user ID",
		}
		utils.RespondWithError(w, apiError)
		return
	}
	userID := int(userIDFloat)
	var requestBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		apiError := models.APIError{
			Code:    models.ErrorCodeBadRequest,
			Message: "Invalid request body",
		}
		utils.RespondWithError(w, apiError)
		return
	}
	if requestBody.Name == "" {
		apiError := models.APIError{
			Code:    models.ErrorCodeBadRequest,
			Message: "Name cannot be empty",
		}
		utils.RespondWithError(w, apiError)
		return
	}
	user := models.User{
		ID:   userID,
		Name: &requestBody.Name,
	}
	updatedUser, err := h.userService.UpdateUser(r.Context(), user)
	if err != nil {
		apiError := models.APIError{
			Code:    models.ErrorCodeInternalServerError,
			Message: "Failed to update user",
		}
		utils.RespondWithError(w, apiError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		apiError := models.APIError{
			Code:    models.ErrorCodeInternalServerError,
			Message: "Error encoding response",
		}
		utils.RespondWithError(w, apiError)
		return
	}
}
