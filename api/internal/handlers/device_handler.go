package handlers

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type DeviceHandler struct {
	deviceService service.DeviceService
	mqttHandler   *MqttHandler // Add this line
}
type assignDeviceRequest struct {
	locationID int `json:"locationID"`
}

// NewDeviceHandler to accept the interface type
// Add mqttHandler to the parameters and assign it
// In handlers/device.go
func NewDeviceHandler(deviceService *service.DefaultDeviceService, mqttHandler *MqttHandler) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		mqttHandler:   mqttHandler, // Add this line
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

func (h *DeviceHandler) GetComponentsByDeviceID(writer http.ResponseWriter, request *http.Request) {
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
	components, err := h.deviceService.GetComponentsByDeviceID(deviceID)
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

func (h *DeviceHandler) MarkComponentLogsAsRead(writer http.ResponseWriter, request *http.Request) {
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
	_, err = h.deviceService.UserHasAccessToComponent(userID, ComponentID)
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
	if err := h.deviceService.MarkComponentLogsAsRead(tx, ComponentID, payload.LogIDs); err != nil {
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

// CreateRecurringSchedule handles the creation of a new recurring schedule.
func (h *DeviceHandler) CreateRecurringSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	var schedule models.RecurringSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", map[string]string{"error": err.Error()}, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}
	log.Printf("Schedules of handler :", schedule)
	// You might want to get the deviceID from the URL vars instead of the body
	vars := mux.Vars(r)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}
	schedule.DeviceID = deviceID

	createdSchedule, err := h.deviceService.CreateRecurringSchedule(r.Context(), &schedule)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to create recurring schedule", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	// once created, call mqtt to send the new schedule to the device
	err = h.mqttHandler.publishSchedules(createdSchedule.DeviceID)
	if err != nil {
		log.Printf("Failed to publish schedules to device %s: %v", createdSchedule.DeviceID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to publish schedule to device", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	utils.RespondWithJSON(w, http.StatusCreated, createdSchedule)
}

// GetRecurringSchedulesByDevice handles retrieving all recurring schedules for a device.
func (h *DeviceHandler) GetRecurringSchedulesByDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	vars := mux.Vars(r)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	schedules, err := h.deviceService.GetRecurringSchedulesByDevice(r.Context(), deviceID)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to get recurring schedules", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, schedules)
}

// UpdateRecurringSchedule handles updating an existing recurring schedule.
func (h *DeviceHandler) UpdateRecurringSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}
	vars := mux.Vars(r)
	scheduleIDStr, ok := vars["scheduleID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing schedule ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid schedule ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	var schedule models.RecurringSchedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid request body", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Call the service to update the schedule
	if err := h.deviceService.UpdateRecurringSchedule(r.Context(), scheduleID, &schedule); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to update recurring schedule", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Schedule updated successfully"})
}

// DeleteRecurringSchedule handles deleting a recurring schedule.
func (h *DeviceHandler) DeleteRecurringSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		apiErr := models.NewAPIError(models.ErrorCodeMethodNotAllowed, "Method not allowed", nil, http.StatusMethodNotAllowed)
		utils.RespondWithError(w, apiErr)
		return
	}

	vars := mux.Vars(r)
	scheduleIDStr, ok := vars["scheduleID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing schedule ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid schedule ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	if err := h.deviceService.DeleteRecurringSchedule(r.Context(), scheduleID); err != nil {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to delete recurring schedule", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	// Update the device with the new schedules
	vars = mux.Vars(r)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}
	err = h.mqttHandler.publishSchedules(deviceID)
	if err != nil {
		log.Printf("Failed to publish schedules to device %s: %v", deviceID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to publish schedule to device", map[string]string{"error": err.Error()}, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DeviceHandler) CheckDevice(id string) bool {
	exists, err := h.deviceService.GetDeviceByDeviceID(id)
	if err != nil {
		log.Printf("Error checking device: %v", err)
		return false
	}
	return exists != nil
}

func (h *DeviceHandler) UpdateDeviceProvisioningToken(id string, token string) bool {
	err := h.deviceService.UpdateDeviceProvisioningToken(id, token)
	if err != nil {
		log.Printf("Error updating device provisioning token: %v", err)
		return false
	}
	return true
}

func (h *DeviceHandler) ProvisionDevice(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceID := vars["deviceID"]
	if deviceID == "" {
		log.Println("ProvisioningDeviceMiddleware: 'deviceID' parameter missing in URL.")
		apiErr := models.NewAPIError(
			models.ErrorCodeMissingParameter,
			"deviceID parameter missing",
			nil,
			http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// 1. Check if the device exists in the database using the device service.
	device, err := h.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		log.Printf("ProvisioningDeviceMiddleware: Error fetching device %s: %v", deviceID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error fetching device", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	if device == nil {
		log.Printf("ProvisioningDeviceMiddleware: Device %s not found", deviceID)
		apiErr := models.NewAPIError(models.ErrorCodeNotFound, "Device not found", nil, http.StatusNotFound)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// 2. Generate a new provisioning token.
	token := utils.GenerateRandomToken(32)
	if token == "" {
		log.Printf("ProvisioningDeviceMiddleware: Error generating token for device %s", deviceID)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error generating token", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// 3. Update the device record with the new token.
	if !h.UpdateDeviceProvisioningToken(deviceID, token) {
		log.Printf("ProvisioningDeviceMiddleware: Error updating token for device %s", deviceID)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error updating token", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	log.Printf("ProvisioningDeviceMiddleware: Successfully updated token for device %s", deviceID)
	// 4. Respond with the provisioning token.
	response := map[string]string{"provisioning_token": token}
	utils.RespondWithJSON(writer, http.StatusOK, response)
}

func (h *DeviceHandler) CheckDeviceRights(token string, id string) bool {
	hasRights, err := h.deviceService.CheckDeviceToken(token, id)
	if err != nil {
		log.Printf("Error checking device rights: %v", err)
		return false
	}
	return hasRights
}

func (h *DeviceHandler) CheckDeviceLocation(id string, id2 string) bool {
	hasRights, err := h.deviceService.CheckDeviceLocation(id, id2)
	if err != nil {
		log.Printf("Error checking device location rights: %v", err)
		return false
	}
	return hasRights
}
