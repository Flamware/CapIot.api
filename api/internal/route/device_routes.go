// internal/handlers/device_handler.go
package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupDeviceRoute initializes the device-related routes using the DeviceHandler
func SetupDeviceRoute(r *mux.Router, deviceHandler *handlers.DeviceHandler) {
	r.Handle("/api/devices", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetAllDevices))).Methods(http.MethodGet)
	r.Handle("/api/unassigned-devices", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetUnassignedDevices))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.DeleteDevice))).Methods(http.MethodDelete)
	r.Handle("/api/unassign-device/{deviceID}", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.UnassignDeviceFromLocation))).Methods(http.MethodPost)
	r.Handle("/api/devices/{deviceID}", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetDeviceByID))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/components", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetcomponentsByDeviceID))).Methods(http.MethodGet)
	// Get all logs for all devices and all components
	r.Handle("/api/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)

	// Get logs for a specific device
	r.Handle("/api/devices/{deviceID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)

	// Get logs for a specific component on a specific device
	r.Handle("/api/devices/{deviceID}/components/{ComponentID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)

	// (Optional) Get logs for a specific component across all devices
	r.Handle("/api/components/{ComponentID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)
	r.Handle("/api/components/{ComponentID}/markAsRead", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.MarkcomponentLogsAsRead))).Methods(http.MethodPost)
	r.Handle("/api/logs/markAllAsRead", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.MarkAllLogsAsRead))).Methods(http.MethodPost)
}
