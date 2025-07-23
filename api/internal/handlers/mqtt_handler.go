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
	mqttClient    mqtt.Client
}

// NewMqttHandler creates a new MqttHandler instance, now taking DeviceService and mqtt.Client
func NewMqttHandler(deviceService *service.DefaultDeviceService, mqttClient mqtt.Client) *MqttHandler {
	return &MqttHandler{
		deviceService: deviceService,
		mqttClient:    mqttClient, // Store the MQTT client
	}
}

// sensorInfo represents the structure of each sensor in the availability message
type sensorInfo struct {
	ID           string  `json:"sensor_id"`
	Type         string  `json:"sensor_type"`
	MinThreshold float64 `json:"min_threshold,omitempty"`
	MaxThreshold float64 `json:"max_threshold,omitempty"`
}

// AvailabilityPayload reflects the structure sent by the Node.js simulator for both availability and LWT
type AvailabilityPayload struct {
	DeviceID  string       `json:"device_id"`
	Status    string       `json:"status"` // Can be "available", "offline", "Running", etc.
	Timestamp string       `json:"timestamp"`
	Sensors   []sensorInfo `json:"sensors,omitempty"` // Sensors might be omitted in LWT "offline" message
}

// AlertPayload reflects the structure sent by the Node.js simulator
type AlertPayload struct {
	DeviceID  string `json:"device_id"`
	SensorID  string `json:"sensor_id"`
	Alert     string `json:"alert"`
	Timestamp string `json:"timestamp"`
}

// HandleDeviceAvailability processes MQTT messages on devices/available/+/ topics.
// This now handles both initial "available" messages and LWT "offline" messages.
func (h *MqttHandler) HandleDeviceAvailability(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received MQTT message on availability topic: %s\n", msg.Topic())

	var payload AvailabilityPayload
	err := json.Unmarshal(msg.Payload(), &payload)
	if err != nil {
		log.Printf("Error unmarshalling JSON for device availability/LWT: %v\n", err)
		return
	}

	deviceID := payload.DeviceID

	if h.deviceService == nil {
		log.Printf("ERROR: h.deviceService is nil for device '%s'!\n", deviceID)
		return
	}

	// --- Step 1: Handle Device Status Update (Offline, Available, etc.) ---
	log.Printf("Processing availability/LWT for device '%s'. Status: '%s'.\n", deviceID, payload.Status)
	err = h.deviceService.UpdateDeviceOperationalStatus(deviceID, models.OperationalStatus(payload.Status))
	if err != nil {
		log.Printf("Error updating device '%s' status to '%s': %v\n", deviceID, payload.Status, err)
		// Even if status update fails, try to proceed with device creation/update if it's "available"
	}

	// If the status is "offline" (from LWT), we don't need to process sensors or link them.
	// The device is simply marked offline.
	if payload.Status == "offline" {
		log.Printf("Device '%s' is offline (LWT/availability update).\n", deviceID)
		// Optionally, you might want to mark all sensors linked to this device as inactive or unlinked.
		// For now, simply updating the device status is sufficient.
		return
	}

	// --- If not offline, proceed with device existence and sensor processing ---
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

	// Always update last seen if device is publishing availability (even if not "Running")
	err = h.deviceService.UpdateDeviceLastSeenAndStatus(deviceID)
	if err != nil {
		log.Printf("Error updating device '%s' last seen: %v\n", deviceID, err)
		// Continue even if last seen update fails
	}

	// --- Step 2: Process and Create/Update sensors (only if sensors are present in payload) ---
	if len(payload.Sensors) > 0 {
		log.Printf("Received availability for device '%s' with %d sensors.\n", payload.DeviceID, len(payload.Sensors))
		for _, sensor := range payload.Sensors {
			var createdOrUpdatedSensor *models.Sensor

			existingSensor, err := h.deviceService.GetsensorByID(sensor.ID)
			if err != nil {
				log.Printf("Error retrieving sensor with ID '%s': %v\n", sensor.ID, err)
				continue
			}

			// Determine actual thresholds, prioritizing existing DB values if not provided by payload (though payload should always provide them here)
			// Or, assume payload provides the desired state and update DB to match
			targetMinThreshold := sensor.MinThreshold
			targetMaxThreshold := sensor.MaxThreshold

			if existingSensor == nil {
				newSensor := &models.Sensor{
					SensorID:     sensor.ID,
					SensorType:   sensor.Type,
					MinThreshold: &targetMinThreshold,
					MaxThreshold: &targetMaxThreshold,
				}
				createdOrUpdatedSensor, err = h.deviceService.Createsensor(newSensor)
				if err != nil {
					log.Printf("Error creating sensor with ID '%s': %v\n", sensor.ID, err)
					continue
				}
				log.Printf("Sensor '%s' (ID: %s) created.\n", createdOrUpdatedSensor.SensorType, createdOrUpdatedSensor.SensorID)
			} else {
				log.Printf("Sensor '%s' (ID: %s) already exists.\n", existingSensor.SensorType, existingSensor.SensorID)
				// Update existing sensor's thresholds if they changed or if payload provides newer values
				updated := false
				if existingSensor.MinThreshold == nil || *existingSensor.MinThreshold != targetMinThreshold {
					existingSensor.MinThreshold = &targetMinThreshold
					updated = true
				}
				if existingSensor.MaxThreshold == nil || *existingSensor.MaxThreshold != targetMaxThreshold {
					existingSensor.MaxThreshold = &targetMaxThreshold
					updated = true
				}
				if updated {
					err = h.deviceService.UpdatesensorRange(payload.DeviceID, existingSensor.SensorID, targetMinThreshold, targetMaxThreshold)
					if err != nil {
						log.Printf("Error updating sensor '%s': %v\n", existingSensor.SensorID, err)
						continue
					}
					log.Printf("Sensor '%s' (ID: %s) updated thresholds to Min: %.2f, Max: %.2f.\n",
						existingSensor.SensorType, existingSensor.SensorID, targetMinThreshold, targetMaxThreshold)
				}
				createdOrUpdatedSensor = existingSensor
			}

			log.Printf("Sensor details (after create/update): ID='%s', Type='%s', MinThreshold=%.2f, MaxThreshold=%.2f\n",
				createdOrUpdatedSensor.SensorID, createdOrUpdatedSensor.SensorType, *createdOrUpdatedSensor.MinThreshold, *createdOrUpdatedSensor.MaxThreshold)

			err = h.deviceService.LinksensorToDevice(deviceID, createdOrUpdatedSensor.SensorID)
			if err != nil {
				// Log the error but don't stop the loop for other sensors
				log.Printf("Error linking sensor '%s' to device '%s': %v\n", createdOrUpdatedSensor.SensorID, deviceID, err)
				continue
			}

			// Re-send config to device if thresholds were just created/updated or need synchronization
			err = h.SetDeviceConfig(deviceID, createdOrUpdatedSensor.SensorID, *createdOrUpdatedSensor.MinThreshold, *createdOrUpdatedSensor.MaxThreshold)
			if err != nil {
				log.Printf("Error re-sending config to device '%s' for sensor '%s': %v\n", deviceID, createdOrUpdatedSensor.SensorID, err)
			} else {
				log.Printf("Resent config to device '%s' for sensor '%s' (Min: %.2f, Max: %.2f) after availability update.\n",
					deviceID, createdOrUpdatedSensor.SensorID, *createdOrUpdatedSensor.MinThreshold, *createdOrUpdatedSensor.MaxThreshold)
			}
		}
	} else {
		log.Printf("No sensors present in availability payload for device '%s'.\n", deviceID)
	}
}

// SetStatus is an HTTP handler that updates the status of a given device via MQTT
func (h *MqttHandler) SetStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to set status for device\n")
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	if deviceID == "" {
		http.Error(w, "Missing deviceID in URL", http.StatusBadRequest)
		return
	}

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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	locationID := 0 // Default value.
	if location == nil {
		// If device is not found, we cannot get a location. This might mean the device isn't registered yet,
		// or the deviceID is invalid.
		log.Printf("Device '%s' not found in DB. Cannot get location.\n", deviceID)
		http.Error(w, "Device not found or not linked to a location", http.StatusNotFound)
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
	token := h.mqttClient.Publish(topic, 0, false, payloadBytes) // QoS 0, not retained
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error: %v", token.Error())
		http.Error(w, "Failed to publish MQTT message", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Status '%s' set for device '%s'", req.Status, deviceID)))
}

// HandleDeviceStatus processes MQTT messages on devices/status/+/ topics.
// This is now primarily for status updates sent by the device itself, NOT LWT.
func (h *MqttHandler) HandleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received status update on topic: %s, message: %s\n", msg.Topic(), string(msg.Payload()))

	parts := strings.Split(msg.Topic(), "/")
	// Expected format: devices/status/deviceID
	if len(parts) != 3 || parts[0] != "devices" || parts[1] != "status" {
		log.Printf("Received message on unexpected status topic format: %s\n", msg.Topic())
		return
	}

	deviceID := parts[2] // Device ID is the third part

	var statusPayload struct {
		DeviceID   string  `json:"device_id"`
		Status     string  `json:"status"`
		LocationID float64 `json:"location_id,omitempty"` // Use float64 for JSON unmarshal
	}

	if err := json.Unmarshal(msg.Payload(), &statusPayload); err != nil {
		log.Printf("Error unmarshalling JSON for device status on topic '%s': %v\n", msg.Topic(), err)
		return
	}

	// Ensure deviceID from topic matches payload, for safety
	if statusPayload.DeviceID != deviceID {
		log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s) for status update on topic %s\n",
			deviceID, statusPayload.DeviceID, msg.Topic())
		// You might choose to return here if a mismatch is critical. For now, we'll proceed with topic's deviceID.
	}

	log.Printf("Received status '%s' from device '%s'.\n", statusPayload.Status, deviceID)

	// Update device's operational status in the database
	err := h.deviceService.UpdateDeviceOperationalStatus(deviceID, models.OperationalStatus(statusPayload.Status))
	if err != nil {
		log.Printf("Error updating device '%s' operational status to '%s': %v", deviceID, statusPayload.Status, err)
		return
	}
	// Also update last seen for general status updates
	err = h.deviceService.UpdateDeviceLastSeenAndStatus(deviceID)
	if err != nil {
		log.Printf("Error updating device '%s' last seen from status update: %v\n", deviceID, err)
	}

	// If device sends "Running" status and has a location_id, send back the config
	if statusPayload.Status == "Running" && statusPayload.LocationID != 0 {
		// Retrieve all sensors linked to this device from your database
		linkedSensors, err := h.deviceService.GetsensorsByDeviceID(deviceID) // Assuming this service method exists
		if err != nil {
			log.Printf("Error getting linked sensors for device '%s': %v", deviceID, err)
			// Still proceed to send back the location_id if no sensors found, but log the issue
		}

		// Iterate through linked sensors and send their config
		for _, sensor := range linkedSensors {
			minThreshold := 0.0
			if sensor.MinThreshold != nil {
				minThreshold = *sensor.MinThreshold
			}
			maxThreshold := 0.0
			if sensor.MaxThreshold != nil {
				maxThreshold = *sensor.MaxThreshold
			}

			// Send individual sensor config to the device
			err = h.SetDeviceConfig(deviceID, sensor.SensorID, minThreshold, maxThreshold)
			if err != nil {
				log.Printf("Error re-sending config for sensor '%s' to device '%s': %v\n", sensor.SensorID, deviceID, err)
			} else {
				log.Printf("Resent config to device '%s' for sensor '%s' (Min: %.2f, Max: %.2f) after 'Running' status.\n",
					deviceID, sensor.SensorID, minThreshold, maxThreshold)
			}
		}

		// Optionally, send a generic "start monitoring" or acknowledgement
		// based on the original Node.js logic that expected a response with location_id.
		// This is a bit redundant if you're sending individual configs, but kept for compatibility.
		payload := map[string]interface{}{
			"device_id":   deviceID,
			"location_id": int(statusPayload.LocationID), // Send as int
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling JSON payload for monitoring start acknowledgement: %v", err)
			return
		}

		// Send back to the specific device's config topic to trigger data monitoring if needed
		// The Node.js simulator listens on its own config topic for this.
		topic := fmt.Sprintf("devices/config/%s", deviceID) // Using config topic for consistency with simulator
		token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
		token.Wait()
		if token.Error() != nil {
			log.Printf("Error publishing MQTT message to acknowledge monitoring start: %v", token.Error())
		} else {
			log.Printf("MQTT acknowledgement published to %s: %s", topic, payloadBytes)
		}
	} else if statusPayload.Status != "Running" {
		// If status is not "Running", ensure location is cleared if applicable
		// (though the Node.js sim handles stopping data monitoring internally)
		log.Printf("Device '%s' status is '%s'. No monitoring start initiated.", deviceID, statusPayload.Status)
	}
}

func (h *MqttHandler) SetDeviceConfig(deviceID string, sensorID string, minThreshold float64, maxThreshold float64) error {
	if deviceID == "" {
		return fmt.Errorf("deviceID cannot be empty")
	}
	if sensorID == "" {
		return fmt.Errorf("sensorID cannot be empty")
	}

	mqttPayload := map[string]interface{}{
		"sensor_id":     sensorID,
		"min_threshold": minThreshold,
		"max_threshold": maxThreshold,
	}

	payloadBytes, err := json.Marshal(mqttPayload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload for config: %v", err)
		return fmt.Errorf("failed to marshal MQTT payload: %w", err)
	}

	topic := "devices/config/" + deviceID
	// QoS 0, not retained for configuration commands (as they are device-specific and current state)
	token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error for config: %v", token.Error())
		return fmt.Errorf("failed to publish MQTT message: %w", token.Error())
	}

	log.Printf("Configuration sent for sensor '%s' on device '%s' with Min: %.2f, Max: %.2f", sensorID, deviceID, minThreshold, maxThreshold)
	return nil
}

func (h *MqttHandler) HandleDeviceAlert(client mqtt.Client, message mqtt.Message) {
	log.Printf("Received alert on topic: %s, message: %s\n", message.Topic(), string(message.Payload()))

	parts := splitTopic(message.Topic())
	// Expected format: devices/alert/deviceID
	if len(parts) != 3 || parts[0] != "devices" || parts[1] != "alert" {
		log.Printf("Received message on unexpected alert topic format: %s (Expected 'devices/alert/deviceID')\n", message.Topic())
		return
	}

	deviceIDFromTopic := parts[2] // Device ID is the third part

	var payload AlertPayload
	err := json.Unmarshal(message.Payload(), &payload)
	if err != nil {
		log.Printf("Error unmarshalling JSON for device alert on topic '%s': %v\n", message.Topic(), err)
		return
	}

	// Validate deviceID from topic matches payload
	if payload.DeviceID != deviceIDFromTopic {
		log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s) for alert on topic %s\n",
			deviceIDFromTopic, payload.DeviceID, message.Topic())
		// Decide if you want to proceed or return here based on strictness
	}

	log.Printf("Received alert '%s' from device '%s' (sensor '%s', timestamp '%s').\n",
		payload.Alert, payload.DeviceID, payload.SensorID, payload.Timestamp)

	// Pass payload data to the service handler
	err = h.deviceService.HandleDeviceAlert(payload.SensorID, payload.Alert) // Assuming this matches your service interface
	if err != nil {
		log.Printf("Error handling device alert for sensor '%s': %v", payload.SensorID, err)
		return
	}
}

// Helper function to split the MQTT topic
func splitTopic(topic string) []string {
	return strings.Split(topic, "/")
}
