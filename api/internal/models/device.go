package models

import "time"

// Component types and subtypes as constants for better type safety and clarity.
const (
	ComponentTypecomponent = "component"
	ComponentTypeActuator  = "actuator"
	ComponentTypeIndicator = "indicator"

	ComponentSubtypeTemperature = "temperature"
	ComponentSubtypeHumidity    = "humidity"
	ComponentSubtypeFan         = "fan"
	ComponentSubtypeLED         = "LED"
)

// Device represents the device data in the database.
type Device struct {
	DeviceID  string    `json:"device_id" db:"device_id"`
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Component represents the component type data.
type Component struct {
	ComponentID      string   `json:"component_id" db:"component_id"`
	ComponentName    string   `json:"component_name" db:"component_name"`
	ComponentType    string   `json:"component_type" db:"component_type"`
	ComponentSubtype string   `json:"component_subtype,omitempty" db:"component_subtype"`
	ComponentStatus  string   `json:"component_status,omitempty" db:"component_status"`
	MinThreshold     *float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold     *float64 `json:"max_threshold,omitempty" db:"max_threshold"`
	MaxRunningHours  *int32   `json:"max_running_hours,omitempty" db:"max_running_hours"`
}

type ComponentRangeUpdate struct {
	ComponentID  string   `json:"component_id" db:"component_id"`
	MinThreshold *float64 `json:"min_threshold,omitempty" db:"min_threshold"`
	MaxThreshold *float64 `json:"max_threshold,omitempty" db:"max_threshold"`
}

// DeviceComponent represents the link between a device and a specific component instance.
type DeviceComponent struct {
	DeviceID         string    `json:"device_id" db:"device_id"`
	ComponentID      string    `json:"component_id" db:"component_id"`
	InstallationDate time.Time `json:"installation_date" db:"installation_date"`
	ExpiryDate       time.Time `json:"expiry_date,omitempty" db:"expiry_date"`
}

// ComponentLog represents the log data from a component.
type ComponentLog struct {
	LogID       int       `json:"log_id" db:"log_id"`
	ComponentID string    `json:"component_id" db:"component_id"`
	Timestamp   time.Time `json:"timestamp" db:"log_timestamp"`
	Content     string    `json:"content" db:"log_content"`
	Read        bool      `json:"read" db:"log_read"`
}

// DeviceLocation represents the link between a device and a location.
type DeviceLocation struct {
	DeviceID   string    `json:"device_id" db:"device_id"`
	LocationID int       `json:"location_id" db:"location_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	IsCurrent  bool      `json:"is_current" db:"is_current"`
}

// DeviceWithComponents represents a device with its associated components.
type DeviceWithComponents struct {
	*Device
	Components []*Component `json:"components,omitempty"`
}

// OperationalStatus represents the operational status of a device.
type OperationalStatus string

const (
	StatusRunning OperationalStatus = "Running"
	StatusStopped OperationalStatus = "Stopped"
	StatusIdle    OperationalStatus = "Idle"
)

// DeviceWithComponentsAndLocation represents a device with its components and location.
type DeviceWithComponentsAndLocation struct {
	*DeviceWithComponents
	Location *Location `json:"location,omitempty"`
}

// AssignDeviceToLocationRequest is used for location assignment via API.
type AssignDeviceToLocationRequest struct {
	DeviceID   string `json:"device_id"`
	LocationID int    `json:"location_id"`
}
