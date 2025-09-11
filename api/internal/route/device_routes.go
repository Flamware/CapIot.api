package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupDeviceRoute initializes the device-related routes using the DeviceHandler
func SetupDeviceRoute(r *mux.Router, deviceHandler *handlers.DeviceHandler, mqttHandler *handlers.MqttHandler) {
	// Public routes that don't need CheckDeviceAccess
	r.Handle("/api/devices", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetAllDevices))).Methods(http.MethodGet)
	r.Handle("/api/unassigned-devices", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetUnassignedDevices))).Methods(http.MethodGet)
	r.Handle("/api/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)
	r.Handle("/api/logs/markAllAsRead", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.MarkAllLogsAsRead))).Methods(http.MethodPost)

	// Routes for a specific device, require CheckDeviceAccess
	r.Handle("/api/devices/{deviceID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.DeleteDevice)))).Methods(http.MethodDelete)
	r.Handle("/api/devices/{deviceID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetDeviceByID)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/components", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetcomponentsByDeviceID)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/logs", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetLogs)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/command", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(mqttHandler.Command)))).Methods(http.MethodPost)
	r.Handle("/api/devices/{deviceID}/sensors", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetSensorsByDeviceID)))).Methods(http.MethodGet)
	// Routes that specify a device and a component
	r.Handle("/api/devices/{deviceID}/components/{componentID}/logs", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetLogs)))).Methods(http.MethodGet)
	// This route has been updated to accept a PATCH method and a simpler URL
	r.Handle("/api/devices/{deviceID}/components/{componentID}/range", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.UpdatecomponentRange)))).Methods(http.MethodPatch)

	// Routes that specify a component and a device
	r.Handle("/api/unassign-device/{deviceID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.UnassignDeviceFromLocation)))).Methods(http.MethodPost)
	r.Handle("/api/components/{componentID}/logs", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetLogs))).Methods(http.MethodGet)
	r.Handle("/api/components/{componentID}/markAsRead", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.MarkcomponentLogsAsRead))).Methods(http.MethodPost)
	r.Handle("/api/components/{componentID}/reset", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(mqttHandler.Reset)))).Methods(http.MethodPost)
}
