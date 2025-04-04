package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"encoding/json"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MqttHandler struct holds the necessary services for processing MQTT messages
type MqttHandler struct {
	deviceService service.DeviceService
	// Add other services as needed
}

// NewMqttHandler creates a new MqttHandler instance
func NewMqttHandler(deviceService service.DeviceService) *MqttHandler {
	return &MqttHandler{
		deviceService: deviceService,
	}
}

// HandleDeviceAvailability processes messages from the "devices/available/+" topic
func (h *MqttHandler) HandleDeviceAvailability(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received MQTT message from topic: %s\n", msg.Topic())

	var device models.Device
	err := json.Unmarshal(msg.Payload(), &device)
	if err != nil {
		log.Printf("Error unmarshalling JSON for device availability: %v\n", err)
		return
	}

	if err := h.deviceService.CreateDevice(&device); err != nil {
		log.Printf("Error creating device via MQTT: %v\n", err)
	}
}

// HandleDeviceData processes messages from the "iot/device/+/data" topic
func (h *MqttHandler) HandleDeviceData(client mqtt.Client, msg mqtt.Message) {
	topicParts := splitTopic(msg.Topic())
	if len(topicParts) > 2 {
		deviceID := topicParts[2]
		payload := msg.Payload()
		log.Printf("Received data for Device ID: %s, Payload: %s\n", deviceID, string(payload))
		// Process the device data using the deviceService or other relevant services
		// Example: h.deviceService.ProcessDeviceData(deviceID, payload)
	} else {
		log.Printf("Invalid device data topic: %s\n", msg.Topic())
	}
}

// Helper function to split the MQTT topic (you might have this elsewhere)
func splitTopic(topic string) []string {
	var parts []string
	for i, r := range topic {
		if r == '/' {
			parts = append(parts, topic[:i])
			topic = topic[i+1:]
		}
	}
	parts = append(parts, topic)
	return parts
}

// Add other handler functions for different MQTT topics as needed
