package models

type Location struct {
	ID          *int    `json:"location_id"`
	Name        *string `json:"location_name"`
	Description *string `json:"location_description"`
}

// NewLocation creates a new Location instance with the provided name and description
func NewLocation(name, description string) *Location {
	return &Location{
		Name:        &name,
		Description: &description,
	}
}

// LocationWithUsersDevices represents a location with its associated users and devices
type LocationWithUsersDevices struct {
	Location
	Users   []*User              `json:"users,omitempty"`   // List of users associated with the location
	Devices []*DeviceWithsensors `json:"devices,omitempty"` // List of devices associated with the location
}

// LocationWithUsers represents a location with its associated users
type LocationWithUsers struct {
	Location
	Users []*User `json:"users,omitempty"` // List of users associated with the location
}

// LocationWithDevices represents a location with its associated devices
type LocationWithDevices struct {
	Location
	Devices []*DeviceWithsensors `json:"devices,omitempty"` // List of devices associated with the location
}

// LocationWithUsersAndDevices represents a location with its associated users and devices
