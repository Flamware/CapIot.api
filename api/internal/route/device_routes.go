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
	r.Handle("/api/devices/{deviceID}/sensors", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetsensorsByDeviceID))).Methods(http.MethodGet)
	// Get all logs for all devices and all sensors
	r.Handle("/api/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)

	// Get logs for a specific device
	r.Handle("/api/devices/{deviceID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)

	// Get logs for a specific sensor on a specific device
	r.Handle("/api/devices/{deviceID}/sensors/{sensorID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)

	// (Optional) Get logs for a specific sensor across all devices
	r.Handle("/api/sensors/{sensorID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)
	r.Handle("/api/sensors/{sensorID}/markAsRead", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.MarkSensorLogsAsRead))).Methods(http.MethodPost)
	r.Handle("/api/logs/markAllAsRead", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.MarkAllLogsAsRead))).Methods(http.MethodPost)
}
