package models

import "time"

// Notification represents a single component log notification for a user.
type Notification struct {
	LogID         int64     `json:"log_id"`         // Primary key in component_log table
	LogRead       bool      `json:"log_read"`       // Whether the notification has been read
	Component     string    `json:"component_name"` // Component name
	ComponentType string    `json:"component_type"` // Component type (sensor, actuator, etc.)
	LogType       string    `json:"log_type"`       // Type of log (info, warning, error, etc.)
	SiteID        int       `json:"site_id"`        // Associated site ID
	SiteName      string    `json:"site_name"`      // Associated site name
	LocationID    int       `json:"location_id"`    // Associated location ID
	LocationName  string    `json:"location_name"`  // Associated location name
	LogTimestamp  time.Time `json:"log_timestamp"`  // Timestamp of the log
	LogContent    string    `json:"log_content"`    // Content/message of the log
}

type Log struct {
	LogID        int64     `json:"log_id"`        // Primary key in component_log table
	ComponentID  int       `json:"component_id"`  // Foreign key to components table
	LogType      string    `json:"log_type"`      // Type of log (info, warning, error, etc.)
	LogTimestamp time.Time `json:"log_timestamp"` // Timestamp of the log
	LogContent   string    `json:"log_content"`   // Content/message of the log
	LogRead      bool      `json:"log_read"`      // Whether the log has been read
}
