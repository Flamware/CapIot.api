package models

import "time"

// Device represents the device data in the database.
type Device struct {
	DeviceID  string    `json:"device_id" db:"device_id"`
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// sensor represents the sensor type data.
type Sensor struct {
	SensorID     string   `json:"sensor_id" db:"sensor_id"`
	SensorType   string   `json:"sensor_type" db:"sensor_type"`
	MinThreshold *float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold *float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

type SensorRangeUpdate struct {
	SensorID     string   `json:"sensor_id" db:"sensor_id"`
	MinThreshold *float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold *float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

// Devicesensor represents the link between a device and a specific sensor instance.
type Devicesensor struct {
	DeviceID     string   `json:"device_id" db:"device_id"`
	SensorID     string   `json:"sensor_id" db:"sensor_id"`
	MinThreshold *float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold *float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

type SensorLog struct {
	SensorID  string    `json:"sensor_id" db:"sensor_id"`
	Timestamp time.Time `json:"timestamp" db:"log_timestamp"`
	Content   string    `json:"content" db:"log_content"`
	Read      bool      `json:"read" db:"log_read"`
	LogID     int       `json:"log_id" db:"log_id"`
}

// DeviceLocation represents the link between a device and a location.
type DeviceLocation struct {
	DeviceID   string    `json:"device_id" db:"device_id"`
	LocationID int       `json:"location_id" db:"location_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	IsCurrent  bool      `json:"is_current" db:"is_current"`
}

// DeviceWithsensors represents a device with its associated sensors.
type DeviceWithsensors struct {
	*Device
	Sensors []*Sensor `json:"sensors,omitempty"`
}

// OperationalStatus represents the operational status of a device.
type OperationalStatus string

const (
	StatusRunning OperationalStatus = "Running"
	StatusStopped OperationalStatus = "Stopped"
	StatusIdle    OperationalStatus = "Idle"
)

// DeviceWithSensorsAndLocation represents a device with its sensors and location.
type DeviceWithSensorsAndLocation struct {
	*DeviceWithsensors
	Location *Location `json:"location,omitempty"`
}

// AssignDeviceToLocationRequest is used for location assignment via API.
type AssignDeviceToLocationRequest struct {
	DeviceID   string `json:"device_id"`
	LocationID int    `json:"location_id"`
}
