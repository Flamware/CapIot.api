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
	r.Handle("/api/assign-device", middleware.JWTAuthMiddleware(http.HandlerFunc(deviceHandler.AssignDeviceToLocation))).Methods(http.MethodPost)
}
