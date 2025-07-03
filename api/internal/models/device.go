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
	CaptorID     string  `json:"captor_id" db:"captor_id"`
	CaptorType   string  `json:"captor_type" db:"captor_type"`
	MinThreshold float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

type CaptorRangeUpdate struct {
	CaptorID     string  `json:"captor_id" db:"captor_id"`
	MinThreshold float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

// DeviceCaptor represents the link between a device and a specific captor instance.
type DeviceCaptor struct {
	DeviceID     string  `json:"device_id" db:"device_id"`
	CaptorID     string  `json:"captor_id" db:"captor_id"`
	MinThreshold float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

// DeviceLocation represents the link between a device and a location.
type DeviceLocation struct {
	DeviceID   string    `json:"device_id" db:"device_id"`
	LocationID int       `json:"location_id" db:"location_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	IsCurrent  bool      `json:"is_current" db:"is_current"`
}

// DeviceWithCaptors represents a device with its associated captors.
type DeviceWithCaptors struct {
	*Device
	Captors []*Captor `json:"captors,omitempty"`
}

// OperationalStatus represents the operational status of a device.
type OperationalStatus string

const (
	StatusRunning OperationalStatus = "Running"
	StatusStopped OperationalStatus = "Stopped"
	StatusIdle    OperationalStatus = "Idle"
)

// DeviceWithSensorsAndLocation represents a device with its captors and location.
type DeviceWithSensorsAndLocation struct {
	*DeviceWithCaptors
	Location *Location `json:"location,omitempty"`
}

// AssignDeviceToLocationRequest is used for location assignment via API.
type AssignDeviceToLocationRequest struct {
	DeviceID   string `json:"device_id"`
	LocationID int    `json:"location_id"`
}
