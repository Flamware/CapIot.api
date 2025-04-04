package route

import (
	service2 "CapIot-api/internal/service"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SetupRouter initializes the routes for the application
func SetupRouter(
	authService *service2.AuthService,
	deviceService *service2.DefaultDeviceService,
	locationService *service2.DefaultLocationService,
	userService *service2.DefaultUserService,
	mqttClient mqtt.Client) *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/api/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a public endpoint"))
	})

	// Health check endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness check endpoint
	mux.HandleFunc("/api/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/api/action", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	SetupAuthRoutes(mux, authService)
	SetupLocationRoutes(mux, locationService)
	SetupDeviceRoute(mux, deviceService)
	SetupUserRoutes(mux, userService)
	SetupMQTTRoutes(mqttClient, deviceService)

	return mux
}
