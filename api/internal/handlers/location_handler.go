package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type LocationHandler struct {
	locationService service.LocationService
}

func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

func (h *LocationHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Ensure Content-Type is application/json
	if r.Header.Get("Content-Type") != "application/json" {
		apiErr := models.NewAPIError(models.ErrorCodeUnsupportedMediaType, "Unsupported Content-Type", nil, http.StatusUnsupportedMediaType)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Decode JSON body into Location struct
	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid JSON", map[string]string{"error": err.Error()}, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Validate input
	if location.Name == nil || *location.Name == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Location name is required", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}
	if location.Description == nil {
		empty := ""
		location.Description = &empty
	}

	log.Printf("Creating location: %+v\n", location)

	// Call service to create location
	if err := h.locationService.CreateLocation(location); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error creating location", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Location created successfully"))
}

func (h *LocationHandler) GetComponentsByLocationID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	vars := mux.Vars(r)
	locationId, ok := vars["locationId"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing location ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	components, err := h.locationService.GetComponentsByLocationID(locationId)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting components", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, components)
}

func (h *LocationHandler) GetDevicesByLocationID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	locationId, ok := vars["locationId"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing location ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	devices, err := h.locationService.GetDevicesByLocationID(locationId)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting devices", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, devices)
}

func (h *LocationHandler) GetAllLocations(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
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
	// Call AuthService to get all locations
	locations, err := h.locationService.GetAllLocations(request.Context(), page, limit, searchTerm)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve data", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, locations)
}

func (h *LocationHandler) CreateSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		apiErr := models.NewAPIError(models.ErrorCodeUnsupportedMediaType, "Unsupported Content-Type", nil, http.StatusUnsupportedMediaType)
		utils.RespondWithError(w, apiErr)
		return
	}

	var site models.Site
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid JSON", map[string]string{"error": err.Error()}, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Validate input
	if site.Name == nil || *site.Name == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Site name is required", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	log.Printf("Creating site: %+v\n", site)

	// Call service to create site
	createdSite, err := h.locationService.CreateSite(r.Context(), site)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error creating site", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, createdSite)
}

// DeleteSite handles DELETE requests to remove a site by its ID.
func (h *LocationHandler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	vars := mux.Vars(r)
	idStr, ok := vars["siteID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing site ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid site ID", map[string]string{"error": err.Error()}, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	if err := h.locationService.DeleteSite(r.Context(), id); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error deleting site", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

func (h *LocationHandler) GetSitesWithPagination(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

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

	// Get the optional search parameter
	searchTerm := r.URL.Query().Get("search")

	sitesData, err := h.locationService.GetSitesWithPagination(r.Context(), page, limit, searchTerm)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve data", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, sitesData)
}

func (h *LocationHandler) GetLocationsBySiteIDs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Read query params
	query := r.URL.Query()
	siteIdsParam := query.Get("site_ids")
	if siteIdsParam == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing site_ids query parameter", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	rawIDs := strings.Split(siteIdsParam, ",")
	var siteIDs []string
	for _, raw := range rawIDs {
		if trimmed := strings.TrimSpace(raw); trimmed != "" {
			siteIDs = append(siteIDs, trimmed)
		}
	}

	// Read page & limit for pagination (optional)
	pageStr := query.Get("page")
	limitStr := query.Get("limit")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10 // default page size
	}

	// Read search term (optional)
	term := query.Get("search")
	// Default to empty string if not provided
	if term == "" {
		term = ""
	}

	// Call service with validated site IDs + user + pagination + search term
	locations, err := h.locationService.GetLocationsBySiteIDs(r.Context(), siteIDs, page, limit, term)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting locations", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, locations)
}

func (h *LocationHandler) CheckSiteAccess(id int, idInt int64) (bool, error) {
	return h.locationService.CheckUserAccessToSite(id, idInt)
}

func (h *LocationHandler) CheckLocationAccess(id int, id2 int64) (bool, error) {
	return h.locationService.CheckUserAccessToLocation(id, id2)
}
