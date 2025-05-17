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
	r.Handle("/api/device/{deviceID}", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetDeviceByID))).Methods(http.MethodGet)
	r.Handle("/api/device/{deviceID}/captors", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.GetCaptorsByDeviceID))).Methods(http.MethodGet)

}
