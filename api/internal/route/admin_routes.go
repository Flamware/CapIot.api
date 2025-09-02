package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/service"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupAdminRoutes sets up the admin-protected route
func SetupAdminRoutes(r *mux.Router,
	adminHandler *handlers.AdminHandler,
	userHandler *handlers.UserHandler,
	locationHandler *handlers.LocationHandler,
	deviceHandler *handlers.DeviceHandler,
	mqttHandler *handlers.MqttHandler,
	authService service.AuthService) {
	adminRouter := r.PathPrefix("/api/admin").Subrouter()

	// Apply JWTAuthMiddleware to all routes under /api/admin
	adminRouter.Use(middleware.JWTAuthMiddleware)

	// Apply RoleCheckMiddleware for "admin" role to all routes under /api/admin
	adminRouter.Use(middleware.RoleCheckMiddleware(authService, "admin"))

	// Admin Test Route
	adminRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin access granted"))
	}).Methods(http.MethodGet)

	// --- User Routes ---
	// Define the handler for updating a user and their locations
	adminRouter.HandleFunc("/users/{userID}", userHandler.UpdateUserAndLocation).Methods(http.MethodPut)
	// Define the handler for getting all the sites associated with a user
	adminRouter.HandleFunc("/users/{userID}/sites", userHandler.GetUserSites).Methods(http.MethodGet)
	// Define the handler for getting users with pagination and search
	adminRouter.HandleFunc("/users", userHandler.GetUsers).Methods(http.MethodGet)

	// --- Site & Location Routes ---
	// Define the handler for getting sites with pagination
	adminRouter.HandleFunc("/sites", locationHandler.GetSitesWithPagination).Methods(http.MethodGet)
	// Define the handler for creating a site
	adminRouter.HandleFunc("/site/create", locationHandler.CreateSite).Methods(http.MethodPost)
	// Define the handler for deleting a site
	adminRouter.HandleFunc("/site/{siteID}", locationHandler.DeleteSite).Methods(http.MethodDelete)
	// Define the handler for getting all locations
	adminRouter.HandleFunc("/locations", locationHandler.GetAllLocations).Methods(http.MethodGet)
	// Define the handler for creating a location
	adminRouter.HandleFunc("/location/create", locationHandler.CreateLocation).Methods(http.MethodPost)
	// Define the handler for modifying a location
	adminRouter.HandleFunc("/location/{locationID}", adminHandler.ModifyLocation).Methods(http.MethodPut)
	// Define the handler for deleting a location
	adminRouter.HandleFunc("/location/{locationID}", adminHandler.DeleteLocation).Methods(http.MethodDelete)

	// --- Device Routes ---
	// Define the handler for getting all devices with their components and locations
	adminRouter.HandleFunc("/devices-components-locations", adminHandler.GetDevicescomponentsLocations).Methods(http.MethodGet)
	// Define the handler for getting all locations with their devices and users
	adminRouter.HandleFunc("/locations-devices-users", adminHandler.GetLocationsDevicesUsers).Methods(http.MethodGet)
	// Define the handler for assigning a device to a location
	adminRouter.HandleFunc("/assign-device", adminHandler.AssignDeviceToLocation).Methods(http.MethodPost)
	// Define the handler for deleting a device
	adminRouter.HandleFunc("/device/{deviceID}", deviceHandler.DeleteDevice).Methods(http.MethodDelete)
	// Define the handler for updating a device's component range
	adminRouter.HandleFunc("/devices/{deviceID}/components/{ComponentID}/range", deviceHandler.UpdatecomponentRange).Methods(http.MethodPut)

	// --- MQTT & Device Status Routes ---
	// Define the handler for starting/stopping monitoring for a device
	adminRouter.HandleFunc("/devices/{deviceID}", mqttHandler.SetStatus).Methods(http.MethodPatch)
	// Define the handler for resetting a device's component timer
	adminRouter.HandleFunc("/devices/{deviceID}/components/{ComponentID}/timer/reset", mqttHandler.ResetTimer).Methods(http.MethodGet)
}
