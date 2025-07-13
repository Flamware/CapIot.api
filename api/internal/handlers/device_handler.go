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

	err = h.deviceService.UpdatesensorRange(DeviceID, SensorID, sensorRangeUpdate.MinThreshold, sensorRangeUpdate.MaxThreshold)
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
