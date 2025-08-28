package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupLocationRoutes initializes the location-related routes
func SetupLocationRoutes(r *mux.Router, locationHandler *handlers.LocationHandler) {
	r.Handle("/api/locations", middleware.JWTAuthMiddleware(http.HandlerFunc(locationHandler.GetAllLocations))).Methods(http.MethodGet)
	r.Handle("/api/location/create", middleware.JWTAuthMiddleware(http.HandlerFunc(locationHandler.CreateLocation))).Methods(http.MethodPost)
	r.Handle("/api/location/{locationId}/components", middleware.JWTAuthMiddleware(http.HandlerFunc(locationHandler.GetComponentsByLocationID))).Methods(http.MethodGet)
	r.Handle("/api/location/{locationId}/devices", middleware.JWTAuthMiddleware(http.HandlerFunc(locationHandler.GetDevicesByLocationID))).Methods(http.MethodGet)
}
