package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"encoding/json"
	"net/http"
)

type LocationHandler struct {
	locationService service.LocationService
}

func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

func (h *LocationHandler) GetAllLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	locations, err := h.locationService.GetAllLocations()
	if err != nil {
		http.Error(w, "Error getting locations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *LocationHandler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := h.locationService.CreateLocation(location); err != nil {
		http.Error(w, "Error creating location", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Location created successfully"))
}
