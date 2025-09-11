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

// NewDeviceHandler to accept the interface type
func NewDeviceHandler(deviceService *service.DefaultDeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

func (h *DeviceHandler) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}
	devices, err := h.deviceService.GetAllDevices()
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error fetching devices", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, devices)
}

func (h *DeviceHandler) DeleteDevice(writer http.ResponseWriter, request *http.Request) {
	log.Printf("Received %s request to delete device", request.Method)

	if request.Method != http.MethodDelete {
		log.Printf("Method %s not allowed...", request.Method)
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		log.Printf("Missing device ID in path")
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	log.Printf("Attempting to delete device with ID: %s", deviceID)

	// Begin transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error starting transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Defer a rollback in case of an error.
	defer tx.Rollback()

	// Pass the transaction to the DeleteDevice method
	if err := h.deviceService.DeleteDevice(tx, deviceID); err != nil {
		log.Printf("Failed to delete device '%s': %v", deviceID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to delete device", map[string]string{"error": err.Error(), "deviceID": deviceID}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Commit the transaction only if all operations were successful
	if err := tx.Commit(); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error committing transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.WriteHeader(http.StatusNoContent)
}
func (h *DeviceHandler) GetDeviceByID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}
	device, err := h.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting device", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	if device == nil {
		apiErr := models.NewAPIError(models.ErrorCodeNotFound, "Device not found", nil, http.StatusNotFound)
		utils.RespondWithError(writer, apiErr)
		return
	}
	utils.RespondWithJSON(writer, http.StatusOK, device)
}

func (h *DeviceHandler) GetcomponentsByDeviceID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}
	components, err := h.deviceService.GetcomponentsByDeviceID(deviceID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting components", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	utils.RespondWithJSON(writer, http.StatusOK, components)
}

func (h *DeviceHandler) GetUnassignedDevices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	unassignedDevices, err := h.deviceService.GetUnassignedDevices()
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting unassigned devices", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	utils.RespondWithJSON(writer, http.StatusOK, unassignedDevices)
}

func (h *DeviceHandler) UnassignDeviceFromLocation(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}
	// Beggin transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error starting transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	defer tx.Rollback()

	if err := h.deviceService.UnassignDeviceFromLocation(tx, deviceID); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error unassigning device from location", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// End transaction
	if err := tx.Commit(); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error committing transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Device unassigned from location successfully"))

}

func (h *DeviceHandler) UpdatecomponentRange(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPatch {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	DeviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	ComponentID, ok := vars["componentID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing component ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Manually read and decode the request body
	var componentRangeUpdate models.ComponentRangeUpdate
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields() // optional but good for catching unknown fields
	err := decoder.Decode(&componentRangeUpdate)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", map[string]string{"error": err.Error()}, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	if componentRangeUpdate.MinThreshold == nil || componentRangeUpdate.MaxThreshold == nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "MinThreshold and MaxThreshold cannot be null", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	err = h.deviceService.UpdateComponentRange(DeviceID, ComponentID, *componentRangeUpdate.MinThreshold, *componentRangeUpdate.MaxThreshold)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to update component range", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
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
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	ComponentID, ok := vars["ComponentID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing component ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	logs, err := h.deviceService.GetcomponentLogsByDeviceIDAndComponentID(deviceID, ComponentID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting component logs", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
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
			apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error fetching logs for specific device and component", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			utils.RespondWithError(w, apiErr)
			return // Important: return after responding to prevent further execution
		}
		utils.RespondWithJSON(w, http.StatusOK, logs)
		return // Important: return after responding to prevent further execution
	}

	if deviceID != "" {
		logs, err = h.deviceService.GetDeviceLogsByDeviceID(deviceID)
		if err != nil {
			apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error fetching all logs for device", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			utils.RespondWithError(w, apiErr)
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, logs)
		return
	}

	if ComponentID != "" {
		logs, err = h.deviceService.GetcomponentLogsByComponentID(ComponentID) // Assuming this method exists in your service
		if err != nil {
			apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error fetching logs for specific component across all devices", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
			utils.RespondWithError(w, apiErr)
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, logs)
		return
	}

	// Get the userId
	userID, err := utils.GetUserIDFromContext(r.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Printf("Retrieved user ID: %s", userID)

	logs, err = h.deviceService.GetAllLogsByUser(userID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error fetching all logs", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, logs)
	return
}

func (h *DeviceHandler) MarkcomponentLogsAsRead(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}

	vars := mux.Vars(request)
	ComponentID, ok := vars["ComponentID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing component ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Define a struct to match the expected JSON payload
	var payload struct {
		LogIDs []int `json:"log_ids"`
	}

	// Decode the JSON payload into the struct
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", map[string]string{"error": err.Error()}, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Check if the component exists
	component, err := h.deviceService.GetComponentByID(ComponentID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting component", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	if component == nil {
		apiErr := models.NewAPIError(models.ErrorCodeNotFound, "component not found", nil, http.StatusNotFound)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Check if the user has access to the component
	userID, err := utils.GetUserIDFromContext(request.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
		utils.RespondWithError(writer, apiErr)
		return
	}
	_, err = h.deviceService.UserHasAccessTocomponent(userID, ComponentID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error checking user access to component", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// beggin transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error starting transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	defer tx.Rollback()

	// Mark the logs as read
	if err := h.deviceService.MarkcomponentLogsAsRead(tx, ComponentID, payload.LogIDs); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error marking component logs as read", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	if err := tx.Commit(); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error committing transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("component logs marked as read successfully")) // Could also use RespondWithJSON for consistency
}

func (h *DeviceHandler) MarkAllLogsAsRead(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}
	// create transaction
	tx, err := h.deviceService.BeginTransaction()
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error starting transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	defer tx.Rollback()
	// Get the userId
	userID, err := utils.GetUserIDFromContext(request.Context())
	if err != nil {
		log.Printf("Failed to retrieve user ID: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
		utils.RespondWithError(writer, apiErr)
		return
	}

	err = h.deviceService.MarkAllLogsAsRead(tx, userID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error marking all logs as read", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	// commit transaction
	if err := tx.Commit(); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error committing transaction", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("All logs marked as read successfully")) // Could also use RespondWithJSON for consistency
}

func (h *DeviceHandler) CheckDeviceAccess(idInt int, id string) bool {
	hasAccess, err := h.deviceService.CheckDeviceAccess(idInt, id)
	if err != nil {
		log.Printf("Error checking device access: %v", err)
		return false
	}
	return hasAccess
}

func (h *DeviceHandler) GetSensorsByDeviceID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(writer, apiErr)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}
	sensors, err := h.deviceService.GetSensorsByDeviceID(deviceID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error getting sensors", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	utils.RespondWithJSON(writer, http.StatusOK, sensors)
}
