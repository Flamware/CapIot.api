package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/service"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupRoleBasedRoutes configures API routes with specific role-based access.
func SetupAdminRoutes(r *mux.Router,
	adminHandler *handlers.AdminHandler,
	userHandler *handlers.UserHandler,
	locationHandler *handlers.LocationHandler,
	deviceHandler *handlers.DeviceHandler,
	mqttHandler *handlers.MqttHandler,
	authService service.AuthService) {

	// A single subrouter for all API routes that require authentication
	apiRouter := r.PathPrefix("/api").Subrouter()

	// Apply JWTAuthMiddleware to all routes under /api
	apiRouter.Use(middleware.JWTAuthMiddleware)

	// --- Routes accessible by 'admin' role ONLY ---
	// Create a sub-subrouter for admin-only routes to avoid repeating middleware
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.RoleCheckMiddleware(authService, []string{"admin"}))

	// Admin Test Route (admin only)
	adminRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin access granted"))
	}).Methods(http.MethodGet)

	// --- User Routes (admin only) ---
	adminRouter.HandleFunc("/users/{userID}", userHandler.UpdateUserAndLocation).Methods(http.MethodPut)
	adminRouter.HandleFunc("/users/{userID}/sites", userHandler.GetUserSites).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users", userHandler.GetUsers).Methods(http.MethodGet)

	// --- Site Routes (admin only) ---
	adminRouter.HandleFunc("/site/create", locationHandler.CreateSite).Methods(http.MethodPost)
	adminRouter.HandleFunc("/site/{siteID}", locationHandler.DeleteSite).Methods(http.MethodDelete)

	// --- Routes accessible by 'admin' and 'operateur' roles ---

	// The `RoleCheckMiddleware` is applied directly to each handler function
	// with a list of allowed roles, giving you fine-grained control.

	// Site Routes
	apiRouter.Handle("/sites",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(locationHandler.GetSitesWithPagination)),
	).Methods(http.MethodGet)

	// Location Routes
	apiRouter.Handle("/locations",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(locationHandler.GetAllLocations)),
	).Methods(http.MethodGet)
	apiRouter.Handle("/location/create",
		middleware.RoleCheckMiddleware(authService, []string{"admin"})(http.HandlerFunc(locationHandler.CreateLocation)),
	).Methods(http.MethodPost)
	apiRouter.Handle("/location/{locationID}",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(adminHandler.ModifyLocation)),
	).Methods(http.MethodPut)
	apiRouter.Handle("/location/{locationID}",
		middleware.RoleCheckMiddleware(authService, []string{"admin"})(http.HandlerFunc(adminHandler.DeleteLocation)),
	).Methods(http.MethodDelete)

	// Device Routes
	apiRouter.Handle("/devices-components-locations",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(adminHandler.GetDevicescomponentsLocations)),
	).Methods(http.MethodGet)
	apiRouter.Handle("/locations-devices-users",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(adminHandler.GetLocationsDevicesUsers)),
	).Methods(http.MethodGet)
	apiRouter.Handle("/assign-device",
		middleware.RoleCheckMiddleware(authService, []string{"admin"})(http.HandlerFunc(adminHandler.AssignDeviceToLocation)),
	).Methods(http.MethodPost)
	apiRouter.Handle("/device/{deviceID}",
		middleware.RoleCheckMiddleware(authService, []string{"admin"})(http.HandlerFunc(deviceHandler.DeleteDevice)),
	).Methods(http.MethodDelete)
	apiRouter.Handle("/devices/{deviceID}/components/{ComponentID}/range",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(deviceHandler.UpdatecomponentRange)),
	).Methods(http.MethodPut)
	apiRouter.Handle("/devices/{deviceID}/components",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(deviceHandler.GetcomponentsByDeviceID)),
	).Methods(http.MethodGet)

	// MQTT & Device Status Routes
	apiRouter.Handle("/devices/{deviceID}",
		middleware.RoleCheckMiddleware(authService, []string{"admin", "operateur"})(http.HandlerFunc(mqttHandler.SetStatus)),
	).Methods(http.MethodPatch)
}
