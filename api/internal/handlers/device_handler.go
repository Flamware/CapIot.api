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

func (h *DeviceHandler) GetcomponentsByDeviceID(writer http.ResponseWriter, request *http.Request) {
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
	components, err := h.deviceService.GetcomponentsByDeviceID(deviceID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting components", map[string]string{"error": err.Error()})
		return
	}
	utils.RespondWithJSON(writer, http.StatusOK, components)
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
	// Beggin transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error starting transaction",
			map[string]string{"error": err.Error()})
		return
	}
	if err := h.deviceService.UnassignDeviceFromLocation(tx, deviceID); err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error unassigning device from location", map[string]string{"error": err.Error()})
		return
	}

	// End transaction
	if err := tx.Commit(); err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error committing transaction",
			map[string]string{"error": err.Error()})
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Device unassigned from location successfully"))

}

func (h *DeviceHandler) UpdatecomponentRange(writer http.ResponseWriter, request *http.Request) {
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

	ComponentID, ok := vars["ComponentID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing component ID in path", nil)
		return
	}

	// Manually read and decode the request body
	var componentRangeUpdate models.ComponentRangeUpdate
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields() // optional but good for catching unknown fields
	err := decoder.Decode(&componentRangeUpdate)
	if err != nil {
		utils.RespondWithError(writer, http.StatusBadRequest, "Invalid request body", map[string]string{"error": err.Error()})
		return
	}

	if componentRangeUpdate.MinThreshold == nil || componentRangeUpdate.MaxThreshold == nil {
		utils.RespondWithError(writer, http.StatusBadRequest, "MinThreshold and MaxThreshold cannot be null", nil)
		return
	}

	err = h.deviceService.UpdateComponentRange(DeviceID, ComponentID, *componentRangeUpdate.MinThreshold, *componentRangeUpdate.MaxThreshold)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Failed to update component range", map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, map[string]string{
		"message":      "component range updated successfully",
		"device_id":    DeviceID,
		"component_id": ComponentID,
	})
}

func (h *DeviceHandler) GetcomponentLogsByDeviceIDAndComponentID(writer http.ResponseWriter, request *http.Request) {
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

	ComponentID, ok := vars["ComponentID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing component ID in path", nil)
		return
	}

	logs, err := h.deviceService.GetcomponentLogsByDeviceIDAndComponentID(deviceID, ComponentID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting component logs", map[string]string{"error": err.Error()})
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, logs)
}

func (h *DeviceHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	// Get parameters from the request path using gorilla/mux.
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	ComponentID := vars["ComponentID"]

	var logs []*models.ComponentLog // Declare a slice to hold the fetched logs
	var err error                   // Declare an error variable

	if deviceID != "" && ComponentID != "" {
		logs, err = h.deviceService.GetcomponentLogsByDeviceIDAndComponentID(deviceID, ComponentID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching logs for specific device and component", map[string]string{"error": err.Error()})
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

	if ComponentID != "" {
		logs, err = h.deviceService.GetcomponentLogsByComponentID(ComponentID) // Assuming this method exists in your service
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error fetching logs for specific component across all devices", map[string]string{"error": err.Error()})
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

func (h *DeviceHandler) MarkcomponentLogsAsRead(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	vars := mux.Vars(request)
	ComponentID, ok := vars["ComponentID"]
	if !ok {
		utils.RespondWithError(writer, http.StatusBadRequest, "Missing component ID in path", nil)
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

	// Check if the component exists
	component, err := h.deviceService.GetComponentByID(ComponentID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error getting component", map[string]string{"error": err.Error()})
		return
	}
	if component == nil {
		utils.RespondWithError(writer, http.StatusNotFound, "component not found", nil)
		return
	}

	// Check if the user has access to the component
	userID, err := utils.GetUserIDFromContext(request.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err = h.deviceService.UserHasAccessTocomponent(userID, ComponentID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error checking user access to component", map[string]string{"error": err.Error()})
		return
	}

	// beggin transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error starting transaction",
			map[string]string{"error": err.Error()})
		return
	}

	// Mark the logs as read
	if err := h.deviceService.MarkcomponentLogsAsRead(tx, ComponentID, payload.LogIDs); err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error marking component logs as read", map[string]string{"error": err.Error()})
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("component logs marked as read successfully")) // Could also use RespondWithJSON for consistency
}

func (h *DeviceHandler) MarkAllLogsAsRead(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		utils.RespondWithError(writer, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}
	// create transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error starting transaction",
			map[string]string{"error": err.Error()})
		return
	}
	// Get the userId
	userID, err := utils.GetUserIDFromContext(request.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		http.Error(writer, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.deviceService.MarkAllLogsAsRead(tx, userID)
	if err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error marking all logs as read", map[string]string{"error": err.Error()})
		return
	}
	// commit transaction
	if err := tx.Commit(); err != nil {
		utils.RespondWithError(writer, http.StatusInternalServerError, "Error committing transaction",
			map[string]string{"error": err.Error()})
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("All logs marked as read successfully")) // Could also use RespondWithJSON for consistency
}
