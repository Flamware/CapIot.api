package models

import "time"

// Device represents the device data in the database.
type Device struct {
	DeviceID  string    `json:"device_id" db:"device_id"`
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Captor represents the captor type data.
type Captor struct {
	CaptorID   string `json:"captor_id" db:"captor_id"`
	CaptorType string `json:"captor_type" db:"captor_type"`
}

// DeviceCaptor represents the link between a device and a specific captor instance.
type DeviceCaptor struct {
	DeviceID string `json:"device_id" db:"device_id"`
	CaptorID string `json:"captor_id" db:"captor_id"`
}
type OperationalStatus string

const (
	StatusOffline OperationalStatus = "offline"
	StatusRunning OperationalStatus = "running"
	StatusStopped OperationalStatus = "stopped"
	// Add other statuses as needed
)
