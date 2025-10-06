package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/service"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
)

// SetupRouter initializes the routes for the application using gorilla/mux
func SetupRouter(
	authHandler *handlers.AuthHandler,
	deviceHandler *handlers.DeviceHandler,
	locationHandler *handlers.LocationHandler,
	userHandler *handlers.UserHandler,
	mqttHandler *handlers.MqttHandler,
	authService service.AuthService, // Add authService as a parameter
	adminHandler *handlers.AdminHandler,
	notificationHandler *handlers.NotificationHandler,
	mqttClient mqtt.Client,
) *mux.Router {
	r := mux.NewRouter()

	// Public endpoint
	r.HandleFunc("/api/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a public endpoint"))
	})

	// Health check endpoint
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness check endpoint
	r.HandleFunc("/api/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	r.HandleFunc("/api/action", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	SetupAuthRoutes(r, authHandler)
	SetupLocationRoutes(r, locationHandler)
	SetupDeviceRoute(r, deviceHandler, mqttHandler)
	SetupUserRoutes(r, userHandler)
	SetupMQTTRoutes(mqttClient, mqttHandler)
	SetupAdminRoutes(r, adminHandler, userHandler, locationHandler, deviceHandler, mqttHandler, authService) // Pass the authService directly
	SetupNotificationsRoute(r, notificationHandler, deviceHandler)

	return r
}
