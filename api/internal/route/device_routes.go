package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupDeviceRoute initializes the device-related routes using the DeviceHandler
func SetupDeviceRoute(r *mux.Router, deviceHandler *handlers.DeviceHandler, mqttHandler *handlers.MqttHandler) {
	// Routes for a specific device, require CheckDeviceAccess
	r.Handle("/api/devices/{deviceID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.DeleteDevice)))).Methods(http.MethodDelete)
	r.Handle("/api/devices/{deviceID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetDeviceByID)))).Methods(http.MethodGet)

	r.Handle("/api/devices/provisioning/{deviceID}", http.HandlerFunc(deviceHandler.ProvisionDevice)).Methods(http.MethodGet)

	r.Handle("/api/devices/check-device-rights/{deviceID}",
		middleware.CheckDeviceRights(deviceHandler)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Access granted"))
			}),
		),
	).Methods(http.MethodGet)

	r.Handle("/api/devices/check-device-location-rights/{deviceID}/{locationID}",
		middleware.CheckDeviceLocationRights(deviceHandler)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Access granted"))
			}),
		),
	).Methods(http.MethodGet)

	r.Handle("/api/devices/check-user-rights/{deviceID}/{locationID}",
		middleware.JWTAuthMiddleware(
			middleware.CheckUserRights(deviceHandler)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Access granted"))
				}),
			),
		),
	).Methods(http.MethodGet)

	r.Handle("/api/devices/check-user-device/{deviceID}",
		middleware.JWTAuthMiddleware(
			middleware.CheckDeviceAccess(deviceHandler)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Access granted"))
				}),
			),
		),
	).Methods(http.MethodGet)

	r.Handle("/api/devices/{deviceID}/components", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetComponentsByDeviceID)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/components/{componentID}/command", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(mqttHandler.HandleCommandComponent)))).Methods(http.MethodPost)
	r.Handle("/api/devices/{deviceID}/sensors", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetSensorsByDeviceID)))).Methods(http.MethodGet)
	// Routes that specify a device and a component
	r.Handle("/api/devices/{deviceID}/components/{componentID}/logs", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetLogs)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/components/{componentID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(mqttHandler.UpdateComponentConfig)))).Methods(http.MethodPatch)

	r.Handle("/api/devices/{deviceID}/command", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(mqttHandler.HandleCommandDevice)))).Methods(http.MethodPost)
	// Routes that specify a component and a device
	r.Handle("/api/unassign-device/{deviceID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.UnassignDeviceFromLocation)))).Methods(http.MethodPost)
	r.Handle("/api/devices/{deviceID}/components/{componentID}/logs", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetLogs)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/components/{componentID}/markAsRead", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.MarkComponentLogsAsRead)))).Methods(http.MethodPost)

	// New routes for scheduling
	// All routes related to schedules should be under a new path, e.g., /schedules
	// They should also require the CheckDeviceAccess middleware

	// Recurring Schedules
	r.Handle("/api/devices/{deviceID}/schedules", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.CreateRecurringSchedule)))).Methods(http.MethodPost)
	r.Handle("/api/devices/{deviceID}/schedules", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.GetRecurringSchedulesByDevice)))).Methods(http.MethodGet)
	r.Handle("/api/devices/{deviceID}/schedules/{scheduleID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.UpdateRecurringSchedule)))).Methods(http.MethodPut)
	r.Handle("/api/devices/{deviceID}/schedules/{scheduleID}", middleware.JWTAuthMiddleware(middleware.CheckDeviceAccess(deviceHandler)(http.HandlerFunc(deviceHandler.DeleteRecurringSchedule)))).Methods(http.MethodDelete)
}
