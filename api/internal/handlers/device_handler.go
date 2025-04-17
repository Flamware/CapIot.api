// internal/handlers/device_handler.go
package handlers

import (
	"CapIot-api/internal/service"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type DeviceHandler struct {
	deviceService service.DeviceService
}
type assignDeviceRequest struct {
	LocationID int `json:"locationID"`
}

// Corrected NewDeviceHandler to accept the interface type
func NewDeviceHandler(deviceService service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

func (h *DeviceHandler) GetAllDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	devices, err := h.deviceService.GetAllDevices()
	if err != nil {
		http.Error(w, "Error fetching devices", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func (h *DeviceHandler) AssignDeviceToLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the LocationID from the URL path
	vars := mux.Vars(r)
	locationIDStr, ok := vars["deviceID"]
	if !ok {
		http.Error(w, "Missing locationId in URL", http.StatusBadRequest)
		return
	}

	var requestData assignDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.deviceService.SetDeviceToLocation(locationIDStr, requestData.LocationID); err != nil {
		http.Error(w, "Error assigning device to location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Device assigned to location successfully"))
}

func (h *DeviceHandler) GetDeviceByID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		http.Error(writer, "Missing device ID in path", http.StatusBadRequest)
		return
	}
	device, err := h.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		http.Error(writer, "Error getting device", http.StatusInternalServerError)
		return
	}
	if device == nil {
		http.Error(writer, "Device not found", http.StatusNotFound)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(device); err != nil {
		http.Error(writer, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *DeviceHandler) GetCaptorsByDeviceID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		http.Error(writer, "Missing device ID in path", http.StatusBadRequest)
		return
	}
	captors, err := h.deviceService.GetCaptorsByDeviceID(deviceID)
	if err != nil {
		http.Error(writer, "Error getting captors", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(captors); err != nil {
		http.Error(writer, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *DeviceHandler) GetUnassignedDevices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	unassignedDevices, err := h.deviceService.GetUnassignedDevices()
	if err != nil {
		http.Error(writer, "Error getting unassigned devices", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK) // Set status code *before* encoding

	if err := json.NewEncoder(writer).Encode(unassignedDevices); err != nil {
		http.Error(writer, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func (h *DeviceHandler) DeleteDevice(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		http.Error(writer, "Missing device ID in path", http.StatusBadRequest)
		return
	}

	if err := h.deviceService.DeleteDevice(deviceID); err != nil {
		http.Error(writer, "Error deleting device", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusNoContent) // No content to return
	writer.Write([]byte("Device deleted successfully"))
	writer.WriteHeader(http.StatusOK) // Set status code *before* encoding
	writer.Write([]byte("Device deleted successfully"))
}

func (h *DeviceHandler) UnassignDeviceFromLocation(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		http.Error(writer, "Missing device ID in path", http.StatusBadRequest)
		return
	}

	if err := h.deviceService.UnassignDeviceFromLocation(deviceID); err != nil {
		http.Error(writer, "Error unassigning device from location", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("Device unassigned from location successfully"))
}
