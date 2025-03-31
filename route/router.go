package route

import (
	"net/http"

	"api.cap.iot/service" // Replace with your actual import path
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SetupRouter initializes the routes for the application
func SetupRouter(authService *service.AuthService, deviceService *service.DefaultDeviceService, mqttClient mqtt.Client) *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a public endpoint"))
	})

	SetupAuthRoutes(mux, authService)
	SetupLocationRoutes(mux)

	// Setup MQTT routes (separate function)
	SetupMQTTRoutes(mqttClient, deviceService) //Passing mqttClient as a parameter

	return mux
}
