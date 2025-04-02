package route

import (
	"api.cap.iot/models"
	"api.cap.iot/service"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
)

func SetupMQTTRoutes(client mqtt.Client, service *service.DefaultDeviceService) {
	topic := "devices/available/+"
	if token := client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message from topic: %s\n", msg.Topic())

		var device models.Device
		err := json.Unmarshal(msg.Payload(), &device)
		if err != nil {
			log.Printf("Error unmarshalling JSON: %v\n", err)
			return
		}
		if err := service.CreateDevice(device); err != nil {
			log.Printf("Error creating device: %v\n", err)
		}
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("Error subscribing to topic: %v", token.Error())
	}
	log.Printf("Subscribed to topic: %s\n", topic)
}
