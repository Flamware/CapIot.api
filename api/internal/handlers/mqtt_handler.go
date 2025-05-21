package handlers

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"

	"CapIot-api/internal/models"  // Import your actual module name
	"CapIot-api/internal/service" // Import your actual module name
)

// MqttHandler struct holds the necessary services for processing MQTT messages
type MqttHandler struct {
	deviceService service.DeviceService
	mqttClient    mqtt.Client // Add MQTT client to the handler.  You removed this, but it's needed.
}

// NewMqttHandler creates a new MqttHandler instance, now taking DeviceService and mqtt.Client
func NewMqttHandler(deviceService *service.DefaultDeviceService, mqttClient mqtt.Client) *MqttHandler {
	return &MqttHandler{
		deviceService: deviceService,
		mqttClient:    mqttClient, // Store the MQTT client
	}
}

// CaptorInfo represents the structure of each captor in the availability message
type AvailabilityPayload struct {
	DeviceID  string       `json:"device_id"`
	Status    string       `json:"status"`
	Timestamp string       `json:"timestamp"`
	Captors   []CaptorInfo `json:"captors"`
}

type CaptorInfo struct {
	ID   string `json:"captor_id"`
	Type string `json:"captor_type"`
}

func (h *MqttHandler) HandleDeviceAvailability(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received MQTT message from topic: %s\n", msg.Topic())

	var payload AvailabilityPayload
	err := json.Unmarshal(msg.Payload(), &payload)
	if err != nil {
		log.Printf("Error unmarshalling JSON for device availability: %v\n", err)
		return
	}

	if h.deviceService == nil {
		log.Printf("ERROR: h.deviceService is nil!")
		return
	}

	deviceID := payload.DeviceID

	// --- Step 1: Ensure Device Exists or Create ---
	device, err := h.deviceService.GetDeviceByDeviceID(deviceID)
	if err != nil {
		log.Printf("Error retrieving device with ID '%s': %v\n", deviceID, err)
		return
	}

	if device == nil {
		newDevice := &models.Device{
			DeviceID: deviceID,
			Status:   payload.Status,
		}
		err = h.deviceService.CreateDevice(newDevice)
		if err != nil {
			log.Printf("Error creating device with ID '%s': %v\n", deviceID, err)
			return
		}
		log.Printf("Device '%s' created.\n", deviceID)
		device = newDevice
	}

	// Update the device's last seen and status
	err = h.deviceService.UpdateDeviceLastSeenAndStatus(deviceID)
	if err != nil {
		log.Printf("Error updating device '%s': %v\n", deviceID, err)
		return
	}

	// --- Step 2: Process and Create Captors and Link ---
	for _, captor := range payload.Captors {
		existingCaptor, err := h.deviceService.GetCaptorByID(captor.ID)
		if err != nil {
			log.Printf("Error retrieving captor with ID '%s': %v\n", captor.ID, err)
			// Consider this a configuration failure
			continue
		}

		var createdCaptor *models.Captor
		if existingCaptor == nil {
			newCaptor := &models.Captor{
				CaptorID:   captor.ID,
				CaptorType: captor.Type,
			}

			createdCaptor, err = h.deviceService.CreateCaptor(newCaptor)
			if err != nil {
				log.Printf("Error creating captor with ID '%s': %v\n", captor.ID, err)
				// Consider this a configuration failure
				continue
			}
			log.Printf("Captor '%s' created with ID '%s'.\n", createdCaptor.CaptorType, createdCaptor.CaptorID)
		} else {
			log.Printf("Captor '%s' already exists with ID '%s'.\n", existingCaptor.CaptorType, existingCaptor.CaptorID)
			createdCaptor = existingCaptor
		}

		if createdCaptor != nil {
			err = h.deviceService.LinkCaptorToDevice(deviceID, createdCaptor.CaptorID)
			if err != nil {
				log.Printf("Error linking captor '%s' to device '%s': %v\n", createdCaptor.CaptorID, deviceID, err)
				// Consider this a configuration failure
				continue
			}
			log.Printf("Captor '%s' linked to device '%s'.\n", createdCaptor.CaptorID, deviceID)
		}
	}
}

// SetStatus is an HTTP handler that updates the status of a given device via MQTT
func (h *MqttHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	// Extract deviceID from the URL path
	log.Printf("Received request to set status for device\n")
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	if deviceID == "" {
		http.Error(w, "Missing deviceID in URL", http.StatusBadRequest)
		return
	}

	// Parse JSON body to get the status
	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		http.Error(w, "Missing status in request body", http.StatusBadRequest)
		return
	}

	//look for location_id in db
	location, err := h.deviceService.GetLocationByDeviceID(deviceID)
	if err != nil {
		log.Printf("Error retrieving location for device '%s': %v\n", deviceID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError) // IMPORTANT: Return error!
		return
	}
	locationID := 0 // Default value.
	if location == nil {
		log.Printf("Device '%s' not found\n", deviceID)
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	if location != nil && location.ID != nil {
		locationID = int(*location.ID)
	}

	// Construct MQTT payload
	payload := map[string]interface{}{
		"device_id": deviceID,
		"status":    req.Status,
	}
	if locationID != 0 {
		payload["location_id"] = locationID
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Publish to MQTT
	topic := "devices/status/" + deviceID
	token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error: %v", token.Error())
		http.Error(w, "Failed to publish MQTT message", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Status set for device: " + deviceID))
}

// HandleDeviceData processes messages from the "iot/device/+/data" topic
func (h *MqttHandler) HandleDeviceData(client mqtt.Client, msg mqtt.Message) {
	topicParts := splitTopic(msg.Topic())
	if len(topicParts) > 2 {
		deviceID := topicParts[2]
		payload := msg.Payload()
		log.Printf("Received data for Device ID: %s, Payload: %s\n", deviceID, string(payload))
	} else {
		log.Printf("Invalid device data topic: %s\n", msg.Topic())
	}
}

func (h *MqttHandler) HandleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received status update on topic: %s, message: %s\n", msg.Topic(), string(msg.Payload()))

	parts := strings.Split(msg.Topic(), "/")
	if len(parts) >= 3 && parts[0] == "devices" && parts[1] == "status" { //check the topic
		deviceID := parts[2] // Device ID is now the third part
		var receivedStatus string
		var locationID float64

		// Process status update payload
		var statusPayload map[string]interface{}
		if err := json.Unmarshal(msg.Payload(), &statusPayload); err == nil {
			if statusValue, ok := statusPayload["status"]; ok {
				if s, ok := statusValue.(string); ok {
					receivedStatus = s
				}
			}
			if locID, ok := statusPayload["location_id"]; ok {
				locationID, _ = locID.(float64) //best effort conversion
			}
		} else {
			receivedStatus = string(msg.Payload()) // Fallback
		}

		log.Printf("Received status '%s' from device '%s'.\n", receivedStatus, deviceID)
		err := h.deviceService.UpdateDeviceOperationalStatus(deviceID, models.OperationalStatus(receivedStatus))
		if err != nil {
			log.Printf("Error updating device status: %v", err)
			return // IMPORTANT:  handle the error
		}

		if receivedStatus == "running" && locationID != 0 {
			payload := map[string]interface{}{
				"device_id":   deviceID,
				"location_id": int(locationID), //send as int
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("Error marshalling JSON payload for monitoring start: %v", err)
				return //  return,  dont send the  mqtt message
			}
			//send mqtt message
			topic := "devices/" + deviceID + "/status" //send back to the specific device topic
			token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
			token.Wait()
			if token.Error() != nil {
				log.Printf("Error publishing MQTT message to start monitoring: %v", token.Error())
			} else {
				log.Printf("MQTT message published to %s: %s", topic, payloadBytes)
			}
		}

	} else if len(parts) >= 3 && parts[0] == "devices" && parts[1] == "lwt" {
		deviceID := parts[2]
		// Process LWT payload (expecting "offline")
		var lwtPayload map[string]interface{}
		receivedStatus := "offline" // Default to offline
		if err := json.Unmarshal(msg.Payload(), &lwtPayload); err == nil {
			if statusValue, ok := lwtPayload["status"]; ok {
				if s, ok := statusValue.(string); ok {
					receivedStatus = s
				}
			}
		} else {
			receivedStatus = string(msg.Payload()) // Fallback
		}
		if receivedStatus == "offline" {
			log.Printf("Device '%s' is offline (LWT).\n", deviceID)
			err := h.deviceService.UpdateDeviceOperationalStatus(deviceID, models.OperationalStatus(receivedStatus))
			if err != nil {
				log.Printf("Error updating device status: %v", err)
				return //handle error
			}
		} else {
			log.Printf("Unexpected LWT status '%s' from device '%s'.\n", receivedStatus, deviceID)
		}
	} else {
		log.Printf("Received message on unexpected base topic: %s\n", msg.Topic())
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
