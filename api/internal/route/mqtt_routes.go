// internal/route/mqtt_routes.go
package route

import (
	"CapIot-api/internal/handlers"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
)

// SetupMQTTRoutes sets up MQTT topic subscriptions and handlers using the MqttHandler
func SetupMQTTRoutes(client mqtt.Client, mqttHandler *handlers.MqttHandler) {
	// Subscribe to the device availability topic and use the handler
	availabilityTopic := "devices/available/+"
	if token := client.Subscribe(availabilityTopic, 1, mqttHandler.HandleDeviceAvailability); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %v", token.Error())
	}
	log.Printf("Subscribed to topic: %s\n", availabilityTopic)

	lwtTopic := "devices/lwt/+"
	if token := client.Subscribe(lwtTopic, 1, mqttHandler.HandleDeviceStatus); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to status topic: %v", token.Error())
	}
	log.Printf("Subscribed to topic: %s\n", lwtTopic)

	statusTopic := "devices/status/+"
	if token := client.Subscribe("devices/status/+", 1, mqttHandler.HandleDeviceStatus); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to status topic: %v", token.Error())
	}

	alertTopic := "devices/alert/+"
	if token := client.Subscribe(alertTopic, 1, mqttHandler.HandleDeviceAlert); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to alert topic: %v", token.Error())
	}

	log.Printf("Subscribed to topic: %s\n", statusTopic)
}
