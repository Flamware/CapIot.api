// internal/route/location_routes.go
package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupLocationRoutes initializes the location-related routes.
// It takes a router, a location handler, and a location service as dependencies.
func SetupLocationRoutes(r *mux.Router, locationHandler *handlers.LocationHandler) {

	// Route for creating a new location.
	// Requires JWT authentication.
	r.Handle("/api/location/create",
		middleware.JWTAuthMiddleware(http.HandlerFunc(locationHandler.CreateLocation)),
	).Methods(http.MethodPost)

	// Route for getting components by location ID.
	// Requires JWT authentication and checks if the user has access to the specified location.
	r.Handle("/api/location/{locationId}/components",
		middleware.JWTAuthMiddleware(middleware.CheckLocationAccess(locationHandler)(http.HandlerFunc(locationHandler.GetComponentsByLocationID))),
	).Methods(http.MethodGet)

	// Route for getting devices by location ID.
	// Requires JWT authentication and checks if the user has access to the specified location.
	r.Handle("/api/location/{locationId}/devices",
		middleware.JWTAuthMiddleware(middleware.CheckLocationAccess(locationHandler)(http.HandlerFunc(locationHandler.GetDevicesByLocationID))),
	).Methods(http.MethodGet)

	// Route for getting locations filtered by site IDs.
	r.Handle("/api/locations/sites",
		middleware.JWTAuthMiddleware(http.HandlerFunc(locationHandler.GetLocationsBySiteIDs)),
	).Methods(http.MethodGet)
}
