package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type LocationHandler struct {
	locationService service.LocationService
}

func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

func (h *AdminHandler) GetAllLocations(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request for all locations", request.Method)

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
	// Call AuthService to get all locations
	locations, err := h.locationService.GetAllLocations(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get all locations: %v", err)
		http.Error(writer, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(locations); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	}
}
func (h *LocationHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure Content-Type is application/json
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
		return
	}

	// Decode JSON body into Location struct
	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate input
	if location.Name == nil || *location.Name == "" {
		http.Error(w, "Location name is required", http.StatusBadRequest)
		return
	}
	if location.Description == nil {
		empty := ""
		location.Description = &empty
	}

	log.Printf("Creating location: %+v\n", location)

	// Call service to create location
	if err := h.locationService.CreateLocation(location); err != nil {
		http.Error(w, "Error creating location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Location created successfully"))
}

func (h *LocationHandler) GetsensorsByLocationID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	locationId, ok := vars["locationId"]
	if !ok {
		http.Error(w, "Missing location ID in path", http.StatusBadRequest)
		return
	}

	sensors, err := h.locationService.GetsensorsByLocationID(locationId)
	if err != nil {
		http.Error(w, "Error getting sensors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sensors); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *LocationHandler) GetDevicesByLocationID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(request)
	locationId, ok := vars["locationId"]
	if !ok {
		http.Error(writer, "Missing location ID in path", http.StatusBadRequest)
		return
	}

	devices, err := h.locationService.GetDevicesByLocationID(locationId)
	if err != nil {
		http.Error(writer, "Error getting devices", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(devices); err != nil {
		http.Error(writer, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *LocationHandler) GetAllLocations(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
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
	// Call AuthService to get all locations
	locations, err := h.locationService.GetAllLocations(request.Context(), page, limit, searchTerm)
	if err != nil {
		log.Printf("Failed to get all locations: %v", err)
		http.Error(writer, "Failed to retrieve data", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(writer).Encode(locations); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(writer, "Failed to send response", http.StatusInternalServerError)
	}
}
