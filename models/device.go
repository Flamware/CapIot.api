package models

// Device represents a device in the system
type Device struct {
	DeviceID   string `json:"device_id"`
	LocationID string `json:"client_id"`
}
