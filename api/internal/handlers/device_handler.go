// internal/handlers/device_handler.go
package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type DeviceHandler struct {
	deviceService service.DeviceService
}
type assignDeviceRequest struct {
	LocationID int `json:"locationID"`
}

// Corrected NewDeviceHandler to accept the interface type
func NewDeviceHandler(deviceService *service.DefaultDeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

func (h *DeviceHandler) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	devices, err := h.deviceService.GetAllDevices()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching devices", map[string]string{"error": err.Error()})
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, devices)
}

func (h *DeviceHandler) DeleteDevice(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to delete device", request.Method)

	if request.Method != http.MethodDelete {
		log.Printf("Method %s not allowed...", request.Method)
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		log.Printf("Missing device ID in path")
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing device ID in path", nil)
		return
	}

	log.Printf("Attempting to delete device with ID: %s", deviceID)

	if err := h.deviceService.DeleteDevice(request.Context(), deviceID); err != nil {
		log.Printf("Failed to delete device '%s': %v", deviceID, err)
		if apiErr, ok := err.(*models.APIError); ok {
			utils.RespondWithError(writer, apiErr.StatusCode, apiErr.Message, apiErr.Details)
			return
		}
		utils.RespondWithError(writer, http.StatusInternalServerError, "Failed to delete device", map[string]string{"error": err.Error(), "deviceID": deviceID})
		return
	}

	writer.WriteHeader(http.StatusNoContent) // Standard response for successful deletion with no body
}

func (h *DeviceHandler) GetDeviceByID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing device ID in path", nil)
		return
	}
	device, err := h.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting device", map[string]string{"error": err.Error()})
		return
	}
	if device == nil {
		utils.RespondWithError(writer, http.StatusNotFound, "Device not found", nil)
		return
	}
	utils.RespondWithJSON(writer, http.StatusOK, device)
}

func (h *DeviceHandler) GetsensorsByDeviceID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing device ID in path", nil)
		return
	}
	sensors, err := h.deviceService.GetsensorsByDeviceID(deviceID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting sensors", map[string]string{"error": err.Error()})
		return
	}
	utils.RespondWithJSON(writer, http.StatusOK, sensors)
}

func (h *DeviceHandler) GetUnassignedDevices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	unassignedDevices, err := h.deviceService.GetUnassignedDevices()
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting unassigned devices", map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, unassignedDevices)
}

func (h *DeviceHandler) UnassignDeviceFromLocation(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing device ID in path", nil)
		return
	}

	if err := h.deviceService.UnassignDeviceFromLocation(deviceID); err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error unassigning device from location", map[string]string{"error": err.Error()})
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Device unassigned from location successfully")) // Could also use RespondWithJSON for consistency
}

func (h *DeviceHandler) UpdatesensorRange(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	DeviceID, ok := vars["deviceID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing device ID in path", nil)
		return
	}

	SensorID, ok := vars["SensorID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing sensor ID in path", nil)
		return
	}

	// Manually read and decode the request body
	var sensorRangeUpdate models.SensorRangeUpdate
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields() // optional but good for catching unknown fields
	err := decoder.Decode(&sensorRangeUpdate)
	if err != nil {
		utils.RespondWithError(writer, http.StatusBadRequest, "Invalid request body", map[string]string{"error": err.Error()})
		return
	}

	if sensorRangeUpdate.MinThreshold == nil || sensorRangeUpdate.MaxThreshold == nil {
		utils.RespondWithError(writer, http.StatusBadRequest, "MinThreshold and MaxThreshold cannot be null", nil)
		return
	}

	err = h.deviceService.UpdatesensorRange(DeviceID, SensorID, *sensorRangeUpdate.MinThreshold, *sensorRangeUpdate.MaxThreshold)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Failed to update sensor range", map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, map[string]string{
		"message":   "sensor range updated successfully",
		"device_id": DeviceID,
		"sensor_id": SensorID,
	})
}

func (h *DeviceHandler) GetSensorLogsByDeviceIDAndSensorID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing device ID in path", nil)
		return
	}

	sensorID, ok := vars["sensorID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing sensor ID in path", nil)
		return
	}

	logs, err := h.deviceService.GetSensorLogsByDeviceIDAndSensorID(deviceID, sensorID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting sensor logs", map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, logs)
}

func (h *DeviceHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// Get parameters from the request path using gorilla/mux.
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	sensorID := vars["sensorID"]

	var logs []*models.SensorLog // Declare a slice to hold the fetched logs
	var err error                // Declare an error variable

	if deviceID != "" && sensorID != "" {
		logs, err = h.deviceService.GetSensorLogsByDeviceIDAndSensorID(deviceID, sensorID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching logs for specific device and sensor", map[string]string{"error": err.Error()})
			return // Important: return after responding to prevent further execution
		}
		utils.RespondWithJSON(w, http.StatusOK, logs)
		return // Important: return after responding to prevent further execution
	}

	if deviceID != "" {
		logs, err = h.deviceService.GetDeviceLogsByDeviceID(deviceID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching all logs for device", map[string]string{"error": err.Error()})
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, logs)
		return
	}

	if sensorID != "" {
		logs, err = h.deviceService.GetSensorLogsBySensorID(sensorID) // Assuming this method exists in your service
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching logs for specific sensor across all devices", map[string]string{"error": err.Error()})
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, logs)
		return
	}

	// Get the userId
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("Retrieved user ID: %s", userID)

	logs, err = h.deviceService.GetAllLogsByUser(userID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching all logs", map[string]string{"error": err.Error()})
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, logs)
	return
}

func (h *DeviceHandler) MarkSensorLogsAsRead(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	sensorID, ok := vars["sensorID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing sensor ID in path", nil)
		return
	}

	// Define a struct to match the expected JSON payload
	var payload struct {
		LogIDs []int `json:"log_ids"`
	}

	// Decode the JSON payload into the struct
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		utils.RespondWithError(writer, http.StatusBadRequest, "Invalid request body", map[string]string{"error": err.Error()})
		return
	}

	// Check if the sensor exists
	sensor, err := h.deviceService.GetsensorByID(sensorID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting sensor", map[string]string{"error": err.Error()})
		return
	}
	if sensor == nil {
		utils.RespondWithError(writer, http.StatusNotFound, "Sensor not found", nil)
		return
	}

	// Check if the user has access to the sensor
	userID, err := utils.GetUserIDFromContext(request.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err = h.deviceService.UserHasAccessToSensor(userID, sensorID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error checking user access to sensor", map[string]string{"error": err.Error()})
		return
	}

	// Mark the logs as read
	if err := h.deviceService.MarkSensorLogsAsRead(sensorID, payload.LogIDs); err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error marking sensor logs as read", map[string]string{"error": err.Error()})
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Sensor logs marked as read successfully")) // Could also use RespondWithJSON for consistency
}

func (h *DeviceHandler) MarkAllLogsAsRead(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get the userId
	userID, err := utils.GetUserIDFromContext(request.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.deviceService.MarkAllLogsAsRead(userID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error marking all logs as read", map[string]string{"error": err.Error()})
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("All logs marked as read successfully")) // Could also use RespondWithJSON for consistency
}
