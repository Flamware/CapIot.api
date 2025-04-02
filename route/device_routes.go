package route

import (
	"api.cap.iot/middleware"
	"api.cap.iot/service"
	"encoding/json"
	"net/http"
)

// SetupDeviceRoute initializes the device-related routes
func SetupDeviceRoute(mux *http.ServeMux, deviceService *service.DefaultDeviceService) {
	mux.Handle("/devices", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		devices, err := deviceService.GetAllDevices()
		if err != nil {
			http.Error(w, "Error fetching devices", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(devices)
	})))

	mux.Handle("/assign-device", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		if err := deviceService.SetDeviceToLocation(requestData.DeviceID, requestData.LocationID); err != nil {
			http.Error(w, "Error assigning device to location", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Device assigned to location successfully"))
	})))
}
