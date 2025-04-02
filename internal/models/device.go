package models

import "time"

// Device represents the simplified device data.
type Device struct {
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
}
