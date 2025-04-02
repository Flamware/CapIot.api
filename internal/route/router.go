package route

import (
	"CapIot-api/internal/service"
	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SetupRouter initializes the routes for the application
func SetupRouter(
	authService *service.AuthService,
	deviceService *service.DefaultDeviceService,
	locationService *service.DefaultLocationService,
	userService *service.DefaultUserService,
	mqttClient mqtt.Client) *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a public endpoint"))
	})

	SetupAuthRoutes(mux, authService)
	SetupLocationRoutes(mux, locationService)
	SetupDeviceRoute(mux, deviceService)
	SetupUserRoutes(mux, userService)
	SetupMQTTRoutes(mqttClient, deviceService)

	return mux
}
