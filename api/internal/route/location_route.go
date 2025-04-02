package route

import (
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"encoding/json"
	"net/http"
)

// SetupLocationRoutes initializes the location-related routes
func SetupLocationRoutes(mux *http.ServeMux, locationService service.LocationService) {
	mux.Handle("/locations", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		locations, err := locationService.GetAllLocations()
		if err != nil {
			http.Error(w, "Error getting locations", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(locations); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
			return
		}
	})))

	mux.Handle("/location/create", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var location models.Location
		if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if err := locationService.CreateLocation(location); err != nil {
			http.Error(w, "Error creating location", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Location created successfully"))
	})))
}
