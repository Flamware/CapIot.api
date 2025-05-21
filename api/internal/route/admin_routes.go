package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/service"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupAdminRoutes sets up the admin-protected route
func SetupAdminRoutes(r *mux.Router, adminHandler *handlers.AdminHandler, userHandler *handlers.UserHandler,
	locationHandler *handlers.LocationHandler,
	mqttHandler *handlers.MqttHandler,
	authService service.AuthService) {
	adminRouter := r.PathPrefix("/api/admin").Subrouter()

	// Apply JWTAuthMiddleware to all routes under /api/admin
	adminRouter.Use(middleware.JWTAuthMiddleware)

	// Apply RoleCheckMiddleware for "admin" role to all routes under /api/admin
	adminRouter.Use(middleware.RoleCheckMiddleware(authService, "admin"))

	// Define the handler for the /test route within the admin subrouter
	adminRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin access granted"))
	}).Methods(http.MethodGet)

	// Define the handler for the /admin/devices-sensors-locations route
	adminRouter.HandleFunc("/devices-sensors-locations", adminHandler.GetDevicesSensorsLocations).Methods(http.MethodGet)

	// Define the handler for the /admin/devices route
	adminRouter.HandleFunc("/users-locations", adminHandler.GetUsersLocations).Methods(http.MethodGet)

	// Define the handler for the /admin/locations-devices-users
	adminRouter.HandleFunc("/locations-devices-users", adminHandler.GetLocationsDevicesUsers).Methods(http.MethodGet)

	// Define the handler for the /admin/locations route
	adminRouter.HandleFunc("/locations", locationHandler.GetAllLocations).Methods(http.MethodGet)

	// Define the handler for assigning a device to a location
	adminRouter.HandleFunc("/assign-device", adminHandler.AssignDeviceToLocation).Methods(http.MethodPost)

	// Define the handler for assigning a user to a location
	adminRouter.HandleFunc("/users/{userID}", userHandler.UpdateUserAndLocation).Methods(http.MethodPut)

	// Define the handler for deleting a device
	adminRouter.HandleFunc("/device/{deviceID}", adminHandler.DeleteDevice).Methods(http.MethodDelete)

	// Define the handler for deleting a location
	adminRouter.HandleFunc("/location/{locationID}", adminHandler.DeleteLocation).Methods(http.MethodDelete)

	// Define the handler for modifying a location
	adminRouter.HandleFunc("/location/{locationID}", adminHandler.ModifyLocation).Methods(http.MethodPut)

	// Define the handler for creating a location
	adminRouter.HandleFunc("/location/create", locationHandler.CreateLocation).Methods(http.MethodPost)

	// Define the handler for starting monitoring for a device
	adminRouter.HandleFunc("/devices/{deviceID}", mqttHandler.SetStatus).Methods(http.MethodPatch)
}
