// internal/handlers/device_handler.go
package handlers

import (
	"CapIot-api/internal/service"
	"encoding/json"
	"net/http"
)

type DeviceHandler struct {
	deviceService service.DeviceService
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

	var requestData struct {
		DeviceID   string `json:"device_id"`
		LocationID int    `json:"location_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.deviceService.SetDeviceToLocation(requestData.DeviceID, requestData.LocationID); err != nil {
		http.Error(w, "Error assigning device to location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Device assigned to location successfully"))
}
