package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"time"
)

// MqttHandler struct holds the necessary services for processing MQTT messages
type MqttHandler struct {
	deviceService service.DeviceService
	mqttClient    mqtt.Client
}

// NewMqttHandler creates a new MqttHandler instance, now taking DeviceService and mqtt.Client
func NewMqttHandler(deviceService service.DeviceService, mqttClient mqtt.Client) *MqttHandler {
	return &MqttHandler{
		deviceService: deviceService,
		mqttClient:    mqttClient,
	}
}

// componentInfo represents the structure of each component in the availability message
type ComponentInfo struct {
	ComponentID      string  `json:"component_id"`
	ComponentName    string  `json:"component_name"`
	ComponentType    string  `json:"component_type"`
	ComponentSubtype string  `json:"component_subtype,omitempty"`
	MinThreshold     float64 `json:"min_threshold,omitempty"`
	MaxThreshold     float64 `json:"max_threshold,omitempty"`
	ComponentStatus  string  `json:"component_status"`
	MaxRunningHours  int     `json:"max_running_hours,omitempty"`
}

// AvailabilityPayload reflects the structure sent by the Node.js simulator for both availability and LWT
type AvailabilityPayload struct {
	DeviceID   string          `json:"device_id"`
	Status     string          `json:"status"`
	Timestamp  string          `json:"timestamp"`
	Components []ComponentInfo `json:"components,omitempty"`
}

// AlertPayload reflects the structure sent by the Node.js simulator
type AlertPayload struct {
	DeviceID    string `json:"device_id"`
	ComponentID string `json:"component_id"`
	Alert       string `json:"alert"`
	Timestamp   string `json:"timestamp"`
}
type ConsumptionPayload struct {
	DeviceID string  `json:"device_id"`
	Voltage  float64 `json:"voltage"`
	Current  float64 `json:"current"`
	Power    float64 `json:"power"`
}

// HandleDeviceAvailability gère le message de disponibilité de l'appareil et effectue le provisionnement
func (h *MqttHandler) HandleDeviceAvailability(client mqtt.Client, msg mqtt.Message) {
	var payload AvailabilityPayload
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		log.Printf("Error unmarshalling availability message: %v", err)
		return
	}

	deviceID := payload.DeviceID
	log.Printf("Received availability message for device: %s", deviceID)

	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	_, err = h.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Device '%s' not found. Creating a new device record.", deviceID)
			newDevice := &models.Device{
				DeviceID:  deviceID,
				Status:    string(models.OperationalStatus(payload.Status)),
				CreatedAt: time.Now(),
			}
			if err = h.deviceService.CreateDevice(tx, newDevice); err != nil {
				log.Printf("Error creating device '%s': %v", deviceID, err)
				return
			}
		} else {
			log.Printf("Error checking for existing device '%s': %v", deviceID, err)
			return
		}
	} else {
		if err = h.deviceService.UpdateDeviceLastSeenAndStatus(tx, deviceID, models.OperationalStatus(payload.Status)); err != nil {
			log.Printf("Error updating device '%s' status: %v", deviceID, err)
		}
	}

	for _, comp := range payload.Components {
		_, getErr := h.deviceService.GetComponentByID(comp.ComponentID)
		if getErr != nil && getErr == sql.ErrNoRows {
			newComponent := &models.Component{
				ComponentID:      comp.ComponentID,
				DeviceID:         deviceID, // Now correctly linking the component to the device
				ComponentName:    comp.ComponentName,
				ComponentType:    comp.ComponentType,
				ComponentSubtype: comp.ComponentSubtype,
				ComponentStatus:  comp.ComponentStatus,
			}

			if comp.MinThreshold != 0 {
				val := comp.MinThreshold
				newComponent.MinThreshold = &val
			}
			if comp.MaxThreshold != 0 {
				val := comp.MaxThreshold
				newComponent.MaxThreshold = &val
			}
			if comp.MaxRunningHours != 0 {
				val := int32(comp.MaxRunningHours)
				newComponent.MaxRunningHours = &val
			}

			if _, err = h.deviceService.CreateComponent(tx, newComponent); err != nil {
				log.Printf("Error creating component '%s': %v", comp.ComponentID, err)
				return
			}
		} else if getErr != nil {
			log.Printf("Error checking for existing component '%s': %v", comp.ComponentID, getErr)
			err = getErr
			return
		}
	}
	if err = h.publishSchedules(deviceID); err != nil {
		log.Printf("Error publishing schedules to device '%s': %v", deviceID, err)
	} else {
		log.Printf("Published schedules to device '%s' successfully.", deviceID)
	}
	linkedComponents, getCompErr := h.deviceService.GetComponentsByDeviceID(deviceID)
	if getCompErr != nil {
		log.Printf("Error getting linked components for device '%s': %v", deviceID, getCompErr)
		err = getCompErr
		return
	}

	for _, component := range linkedComponents {
		minThreshold := 0.0
		if component.MinThreshold != nil {
			minThreshold = *component.MinThreshold
		}
		maxThreshold := 0.0
		if component.MaxThreshold != nil {
			maxThreshold = *component.MaxThreshold
		}

		if err = h.publishConfigToDevice(deviceID, component.ComponentID, &minThreshold, &maxThreshold, component.MaxRunningHours); err != nil {
			log.Printf("Error re-sending config for component '%s' to device '%s': %v\n", component.ComponentID, deviceID, err)
		} else {
			log.Printf("Resent config to device '%s' for component '%s' (Min: %.2f, Max: %.2f) after 'Running' status.\n",
				deviceID, component.ComponentID, minThreshold, maxThreshold)
		}
	}
	log.Printf("Device '%s' availability processed successfully.", deviceID)
}

// HandleDeviceStatus handles status updates sent by the device itself
func (h *MqttHandler) HandleDeviceStatus(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received status update on topic: %s, message: %s\n", msg.Topic(), string(msg.Payload()))

	parts := strings.Split(msg.Topic(), "/")
	if len(parts) != 3 || parts[0] != "devices" || parts[1] != "status" {
		log.Printf("Received message on unexpected status topic format: %s\n", msg.Topic())
		return
	}

	deviceID := parts[2]

	var statusPayload struct {
		DeviceID   string  `json:"device_id"`
		Status     string  `json:"status"`
		locationID float64 `json:"location_id,omitempty"`
	}

	if err := json.Unmarshal(msg.Payload(), &statusPayload); err != nil {
		log.Printf("Error unmarshalling JSON for device status on topic '%s': %v\n", msg.Topic(), err)
		return
	}

	if statusPayload.DeviceID != deviceID {
		log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s) for status update on topic %s\n",
			deviceID, statusPayload.DeviceID, msg.Topic())
	}

	log.Printf("Received status '%s' from device '%s'.\n", statusPayload.Status, deviceID)

	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction for status update: %v", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	// if status is offline update the consumption to null
	if statusPayload.Status == "offline" {
		log.Printf("Device '%s' is offline. Setting consumption values to null.\n", deviceID)
		// Update consumption values to null
		if err = h.deviceService.UpdateDeviceConsumption(tx, deviceID, nil, nil, nil); err != nil {
			log.Printf("Error updating consumption to null for device '%s': %v", deviceID, err)
			return
		}
	}

	if err = h.deviceService.UpdateDeviceOperationalStatus(tx, deviceID, models.OperationalStatus(statusPayload.Status)); err != nil {
		log.Printf("Error updating device '%s' operational status to '%s': %v", deviceID, statusPayload.Status, err)
		return
	}
	if err = h.deviceService.UpdateDeviceLastSeenAndStatus(tx, deviceID, models.OperationalStatus(statusPayload.Status)); err != nil {
		log.Printf("Error updating device '%s' last seen from status update: %v\n", deviceID, err)
		return
	}
}

// HandleDeviceAlert handles incoming alert messages from devices.
func (h *MqttHandler) HandleDeviceAlert(client mqtt.Client, message mqtt.Message) {
	log.Printf("Received alert on topic: %s, message: %s\n", message.Topic(), string(message.Payload()))

	parts := splitTopic(message.Topic())
	if len(parts) != 3 || parts[0] != "devices" || parts[1] != "alert" {
		log.Printf("Received message on unexpected alert topic format: %s (Expected 'devices/alert/deviceID')\n", message.Topic())
		return
	}

	deviceIDFromTopic := parts[2]

	var payload AlertPayload
	err := json.Unmarshal(message.Payload(), &payload)
	if err != nil {
		log.Printf("Error unmarshalling JSON for device alert on topic '%s': %v\n", message.Topic(), err)
		return
	}

	if payload.DeviceID != deviceIDFromTopic {
		log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s) for alert on topic %s\n",
			deviceIDFromTopic, payload.DeviceID, message.Topic())
	}

	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction for alert: %v", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err = h.deviceService.HandleDeviceAlert(tx, payload.ComponentID, payload.Alert); err != nil {
		log.Printf("Error handling device alert for component '%s': %v", payload.ComponentID, err)
		return
	}
}

// HandleRunningHours handles incoming running hours updates from devices.
func (h *MqttHandler) HandleRunningHours(client mqtt.Client, message mqtt.Message) {
	log.Printf("Received running hours update on topic: %s, message: %s\n", message.Topic(), string(message.Payload()))

	parts := splitTopic(message.Topic())
	if len(parts) != 3 || parts[0] != "devices" || parts[1] != "running_hours" {
		log.Printf("Received message on unexpected running hours topic format: %s\n", message.Topic())
		return
	}

	deviceID := parts[2]

	var payload struct {
		DeviceID    string `json:"device_id"`
		ComponentID string `json:"component_id"`
		Hours       int32  `json:"running_hours"`
		Timestamp   string `json:"timestamp"`
	}

	if err := json.Unmarshal(message.Payload(), &payload); err != nil {
		log.Printf("Error unmarshalling JSON for running hours on topic '%s': %v\n", message.Topic(), err)
		return
	}

	if payload.DeviceID != deviceID {
		log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s) for running hours update on topic %s\n",
			deviceID, payload.DeviceID, message.Topic())
	}

	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction for running hours update: %v", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	if err = h.deviceService.UpdateComponentRunningHours(tx, payload.ComponentID, payload.Hours); err != nil {
		log.Printf("Error updating running hours for component '%s': %v", payload.ComponentID, err)
		return
	}
}

// HandleCommandDevice handles commands for an entire device (e.g., Start, Stop).
func (d *MqttHandler) HandleCommandDevice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	if deviceID == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing deviceID in URL", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid JSON payload", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}
	if req.Command == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing 'command' in request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	payload := map[string]interface{}{
		"device_id": deviceID,
		"command":   req.Command,
	}
	responseMsg := ""

	switch req.Command {
	case "Start":
		location, err := d.deviceService.GetLocationByDeviceID(deviceID)
		if err != nil {
			if err == sql.ErrNoRows {
				apiErr := models.NewAPIError(models.ErrorCodeNotFound, "Device not found or not linked to a location", nil, http.StatusNotFound)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("Error retrieving location for device '%s': %v\n", deviceID, err)
			apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal server error", nil, http.StatusInternalServerError)
			utils.RespondWithError(w, apiErr)
			return
		}
		if location != nil && location.ID != nil {
			payload["location_id"] = *location.ID
		}
		responseMsg = fmt.Sprintf("Command '%s' sent to device '%s'", req.Command, deviceID)

	case "Stop":
		responseMsg = fmt.Sprintf("Command '%s' sent to device '%s'", req.Command, deviceID)
	case "Follow_Schedule":
		responseMsg = fmt.Sprintf("Command '%s' sent to device '%s'", req.Command, deviceID)
	default:
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid command. Supported commands are 'Start' and 'Stop'.", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal server error", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	topic := "devices/commands/" + deviceID
	token := d.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error: %v", token.Error())
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to publish MQTT message", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Printf("MQTT command '%s' published to %s: %s", req.Command, topic, payloadBytes)

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": responseMsg})
}

// HandleCommandComponent handles commands for a specific device component (e.g., reset, set_hours, update_config).
func (d *MqttHandler) HandleCommandComponent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	componentID := vars["componentID"]

	if deviceID == "" || componentID == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing deviceID or componentID in URL", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	var req struct {
		Command         string   `json:"command"`
		Hours           *int32   `json:"hours,omitempty"`
		MinThreshold    *float64 `json:"min_threshold,omitempty"`
		MaxThreshold    *float64 `json:"max_threshold,omitempty"`
		MaxRunningHours *int32   `json:"max_running_hours,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid JSON payload", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	if req.Command == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing 'command' in request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	var mqttPayload map[string]interface{}
	responseMsg := ""

	tx, err := d.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal server error", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	switch req.Command {
	case "reset":
		if err = d.deviceService.ResetComponentRunningHours(tx, componentID); err != nil {
			log.Printf("Error resetting running hours for component '%s': %v", componentID, err)
			apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to reset running hours", nil, http.StatusInternalServerError)
			utils.RespondWithError(w, apiErr)
			return
		}
		mqttPayload = map[string]interface{}{
			"command":      req.Command,
			"component_id": componentID,
		}
		responseMsg = fmt.Sprintf("Command '%s' for component '%s' sent to device '%s'", req.Command, componentID, deviceID)

	default:
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid command. Supported commands are 'reset', 'set_hours', and 'update_config'.", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	payloadBytes, err := json.Marshal(mqttPayload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal server error", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	topic := "devices/commands/" + deviceID
	token := d.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error: %v", token.Error())
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to publish MQTT message", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Printf("MQTT command '%s' published to %s: %s", req.Command, topic, payloadBytes)

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": responseMsg})
}

// UpdateComponentConfig handles PATCH requests to update a component's configuration.
func (h *MqttHandler) UpdateComponentConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	componentID := vars["componentID"]

	if deviceID == "" || componentID == "" {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing deviceID or componentID in URL", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	var req struct {
		MinThreshold    *float64 `json:"min_threshold,omitempty"`
		MaxThreshold    *float64 `json:"max_threshold,omitempty"`
		MaxRunningHours *int32   `json:"max_running_hours,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid JSON payload", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	if req.MinThreshold == nil && req.MaxThreshold == nil && req.MaxRunningHours == nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "At least one of 'min_threshold', 'max_threshold', or 'max_running_hours' must be provided", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction for config update: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal server error", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	newConfig := models.ComponentConfig{
		ComponentID:     componentID,
		MinThreshold:    req.MinThreshold,
		MaxThreshold:    req.MaxThreshold,
		MaxRunningHours: req.MaxRunningHours,
	}

	if err := h.deviceService.UpdateComponentConfig(tx, newConfig); err != nil {
		log.Printf("Error updating component '%s' config in database: %v", componentID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to update component configuration", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	if err := h.publishConfigToDevice(deviceID, componentID, req.MinThreshold, req.MaxThreshold, req.MaxRunningHours); err != nil {
		log.Printf("Error publishing config to device '%s': %v", deviceID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Configuration updated in database, but failed to publish to device.", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Configuration updated successfully and published to device."})
}

// publishConfigToDevice is a helper function to send configuration updates via MQTT.
func (h *MqttHandler) publishConfigToDevice(deviceID, componentID string, minThreshold, maxThreshold *float64, maxRunningHours *int32) error {
	if deviceID == "" || componentID == "" {
		return fmt.Errorf("deviceID and componentID cannot be empty")
	}

	// Prepare the payload, including only the non-nil fields.
	mqttPayload := map[string]interface{}{
		"component_id": componentID,
	}
	if minThreshold != nil {
		mqttPayload["min_threshold"] = *minThreshold
	}
	if maxThreshold != nil {
		mqttPayload["max_threshold"] = *maxThreshold
	}
	if maxRunningHours != nil {
		mqttPayload["max_running_hours"] = *maxRunningHours
	}

	payloadBytes, err := json.Marshal(mqttPayload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload for config: %v", err)
		return fmt.Errorf("failed to marshal MQTT payload: %w", err)
	}

	topic := fmt.Sprintf("devices/config/%s", deviceID)
	token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error for config: %v", token.Error())
		return fmt.Errorf("failed to publish MQTT message: %w", token.Error())
	}

	log.Printf("Configuration sent to device '%s' for component '%s'.", deviceID, componentID)
	return nil
}

// Helper function to split the MQTT topic
func splitTopic(topic string) []string {
	return strings.Split(topic, "/")
}

// publishSchedules is a helper function to send schedule updates via MQTT.
func (h *MqttHandler) publishSchedules(deviceID string) error {
	if deviceID == "" {
		return fmt.Errorf("deviceID cannot be empty")
	}

	schedules, err := h.deviceService.GetRecurringSchedulesByDevice(context.Background(), deviceID)
	if err != nil {
		log.Printf("Error retrieving schedules for device '%s': %v", deviceID, err)
		return fmt.Errorf("failed to retrieve schedules: %w", err)
	}

	payload := map[string]interface{}{
		"schedules": schedules,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling MQTT payload for schedules: %v", err)
		return fmt.Errorf("failed to marshal MQTT payload: %w", err)
	}

	topic := fmt.Sprintf("devices/schedules/%s", deviceID)
	token := h.mqttClient.Publish(topic, 0, false, payloadBytes)
	token.Wait()
	if token.Error() != nil {
		log.Printf("MQTT publish error for schedules: %v", token.Error())
		return fmt.Errorf("failed to publish MQTT message: %w", token.Error())
	}

	log.Printf("Schedules sent to device '%s'.", deviceID)
	return nil
}

// This func handles the data of : current, voltage and power sent per a device
func (h *MqttHandler) HandleConsumption(client mqtt.Client, message mqtt.Message) {
	log.Printf("Received consumption update on topic: %s, message: %s\n", message.Topic(), string(message.Payload()))

	// Expected topic format: devices/consumption/<deviceID>
	parts := strings.Split(message.Topic(), "/")
	if len(parts) != 3 || parts[0] != "devices" || parts[1] != "consumption" {
		log.Printf("Received message on unexpected consumption topic format: %s\n", message.Topic())
		return
	}

	var payload ConsumptionPayload
	if err := json.Unmarshal(message.Payload(), &payload); err != nil {
		log.Printf("Error unmarshalling JSON for consumption on topic '%s': %v\n", message.Topic(), err)
		return
	}

	// Ensure the payload device matches the topic device
	if payload.DeviceID != parts[2] {
		log.Printf("Warning: DeviceID mismatch between topic (%s) and payload (%s)\n", parts[2], payload.DeviceID)
	}

	// Begin transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		log.Printf("Error beginning transaction for consumption update: %v", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Update the device consumption in DB
	if err = h.deviceService.UpdateDeviceConsumption(tx, payload.DeviceID, &payload.Current, &payload.Voltage, &payload.Power); err != nil {
		log.Printf("Error updating consumption for device '%s': %v", payload.DeviceID, err)
		return
	}

	log.Printf("Consumption updated successfully for device '%s'", payload.DeviceID)
}
