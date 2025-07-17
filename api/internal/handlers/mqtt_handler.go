package handlers

import (
	"encoding/json"
	"fmt"
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

type sensorInfo struct {
	ID           string  `json:"sensor_id"`
	Type         string  `json:"sensor_type"`
	MinThreshold float64 `json:"min_threshold,omitempty"`
	MaxThreshold float64 `json:"max_threshold,omitempty"`
}

// sensorInfo represents the structure of each sensor in the availability message
type AvailabilityPayload struct {
	DeviceID  string       `json:"device_id"`
	Status    string       `json:"status"`
	Timestamp string       `json:"timestamp"`
	Sensors   []sensorInfo `json:"sensors"`
}

// AlertPayload reflects the structure sent by the Node.js simulator
type AlertPayload struct {
	DeviceID  string `json:"device_id"`
	SensorID  string `json:"sensor_id"`
	Alert     string `json:"alert"`
	Timestamp string `json:"timestamp"`
}

func (h *MqttHandler) HandleDeviceAvailability(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received MQTT message from topic: %s\n", msg.Topic())

	var payload AvailabilityPayload
	err := json.Unmarshal(msg.Payload(), &payload)
	if err != nil {
		log.Printf("Error unmarshalling JSON for device availability: %v\n", err)
		return
	}
	log.Printf("Received availability for device '%s' with sensors '%v'.\n", payload.DeviceID, payload.Sensors) // Corrected log format

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

	err = h.deviceService.UpdateDeviceLastSeenAndStatus(deviceID)
	if err != nil {
		log.Printf("Error updating device '%s': %v\n", deviceID, err)
		return
	}

	// --- Step 2: Process and Create/Update sensors ---
	for _, sensor := range payload.Sensors {
		var createdOrUpdatedSensor *models.Sensor // Use a more descriptive name

		existingSensor, err := h.deviceService.GetsensorByID(sensor.ID) // Corrected variable name
		if err != nil {
			log.Printf("Error retrieving sensor with ID '%s': %v\n", sensor.ID, err)
			continue
		}

		if existingSensor == nil {
			newSensor := &models.Sensor{ // Corrected variable name
				SensorID:     sensor.ID,
				SensorType:   sensor.Type,
				MinThreshold: &sensor.MinThreshold,
				MaxThreshold: &sensor.MaxThreshold,
			}
			createdOrUpdatedSensor, err = h.deviceService.Createsensor(newSensor) // Corrected variable name
			if err != nil {
				log.Printf("Error creating sensor with ID '%s': %v\n", sensor.ID, err)
				continue
			}
			log.Printf("Sensor '%s' created with ID '%s'.\n", createdOrUpdatedSensor.SensorType, createdOrUpdatedSensor.SensorID)
		} else {
			log.Printf("Sensor '%s' already exists with ID '%s'.\n", existingSensor.SensorType, existingSensor.SensorID)
			createdOrUpdatedSensor = existingSensor
		}

		// Print all sensor details for debugging
		log.Printf("Sensor details: ID='%s', Type='%s', MinThreshold=%.2f, MaxThreshold=%.2f\n",
			createdOrUpdatedSensor.SensorID, createdOrUpdatedSensor.SensorType,
			createdOrUpdatedSensor.MinThreshold, createdOrUpdatedSensor.MaxThreshold)

		err = h.deviceService.LinksensorToDevice(deviceID, createdOrUpdatedSensor.SensorID)
		if err != nil {
			log.Printf("Error linking sensor '%s' to device '%s': %v\n", createdOrUpdatedSensor.SensorID, deviceID, err)
			continue
		}

		// Use the thresholds that were just confirmed or updated in the database/application's state
		err = h.SetDeviceConfig(deviceID, createdOrUpdatedSensor.SensorID, *createdOrUpdatedSensor.MinThreshold, *createdOrUpdatedSensor.MaxThreshold)
		if err != nil {
			log.Printf("Error re-sending config to device '%s' for sensor '%s': %v\n", deviceID, createdOrUpdatedSensor.SensorID, err)
		} else {
			log.Printf("Resent config to device '%s' for sensor '%s' (Min: %.2f, Max: %.2f)\n",
				deviceID, createdOrUpdatedSensor.SensorID, createdOrUpdatedSensor.MinThreshold, createdOrUpdatedSensor.MaxThreshold)
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

func (h *MqttHandler) SetDeviceConfig(deviceID string, sensorID string, minThreshold float64, maxThreshold float64) error {
	if deviceID == "" {
		return fmt.Errorf("deviceID cannot be empty")
	}
	if sensorID == "" {
		return fmt.Errorf("sensorID cannot be empty")
	}

	// Construct the payload to be sent via MQTT
	// This payload now explicitly includes sensor_id and its specific thresholds
	mqttPayload := map[string]interface{}{
		"sensor_id":     sensorID,
		"min_threshold": minThreshold,
		"max_threshold": maxThreshold,
	}

	payloadBytes, err := json.Marshal(mqttPayload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload: %v", err)
		return fmt.Errorf("failed to marshal MQTT payload: %w", err)
	}

	// The topic remains devices/config/deviceID as per your request,
	// with sensor-specific configuration now inside the payload.
	topic := "devices/config/" + deviceID
	token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait() // Wait for the publish operation to complete
	if token.Error() != nil {
		log.Printf("MQTT publish error: %v", token.Error())
		return fmt.Errorf("failed to publish MQTT message: %w", token.Error())
	}

	log.Printf("Configuration sent for sensor '%s' on device '%s' with Min: %.2f, Max: %.2f", sensorID, deviceID, minThreshold, maxThreshold)
	return nil // Return nil on success
}

func (h *MqttHandler) HandleDeviceAlert(client mqtt.Client, message mqtt.Message) {
	log.Printf("Received alert on topic: %s, message: %s\n", message.Topic(), string(message.Payload()))

	parts := splitTopic(message.Topic())
	log.Println("Split topic parts:", parts) // Debugging log to see the split parts
	// Ensure there are at least 3 parts for "devices/alert/deviceID"
	if len(parts) >= 3 && parts[0] == "devices" && parts[1] == "alert" {
		deviceIDFromTopic := parts[2] // Device ID is the third part

		var payload AlertPayload
		err := json.Unmarshal(message.Payload(), &payload)
		if err != nil {
			log.Printf("Error unmarshalling JSON for device alert on topic '%s': %v\n", message.Topic(), err)
			return
		}

		// Validate deviceID from topic matches payload
		if payload.DeviceID != deviceIDFromTopic { // Use deviceIDFromTopic here
			log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s) for alert on topic %s\n",
				deviceIDFromTopic, payload.DeviceID, message.Topic())
		}

		// Updated log message to reflect removed fields
		log.Printf("Received alert '%s' from device '%s' (sensor '%s', timestamp '%s').\n",
			payload.Alert, payload.DeviceID, payload.SensorID, payload.Timestamp)

		// Pass payload data to the service handler
		// Note: The service layer's HandleDeviceAlert might still expect 'type' or 'value'
		// If so, you'll need to adapt the service interface/implementation or populate
		// these fields based on `sensor_id` lookup in your DB. For now, matching the direct payload.
		err = h.deviceService.HandleDeviceAlert(payload.SensorID, payload.Alert)
		if err != nil {
			log.Printf("Error handling device alert: %v", err)
			return // IMPORTANT: handle the error
		}
	} else {
		// Log a more informative message if the topic format is unexpected
		log.Printf("Received message on unexpected topic format for alert: %s (Expected 'devices/alert/deviceID')\n", message.Topic())
	}
}

// Helper function to split the MQTT topic (you might have this elsewhere)
func splitTopic(topic string) []string {
	// This is the correct and idiomatic way to split a string by a delimiter in Go
	return strings.Split(topic, "/")
}
