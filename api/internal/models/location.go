package models

// Site represents a physical site or campus.
type Site struct {
	ID      *int    `json:"site_id"`
	Name    *string `json:"site_name"`
	Address *string `json:"site_address"`
}

// Location represents a specific location within a Site.
type Location struct {
	ID          *int    `json:"location_id"`
	Name        *string `json:"location_name"`
	Description *string `json:"location_description"`
	SiteID      *int    `json:"site_id"`
}

// NewLocation creates a new Location instance with the provided details.
func NewLocation(name, description string, siteID int, siteName string) *Location {
	return &Location{
		Name:        &name,
		Description: &description,
		SiteID:      &siteID,
	}
}

// LocationWithUsersDevices represents a location with its associated users and devices.
type LocationWithUsersDevices struct {
	Location
	Users   []*User                 `json:"users,omitempty"`   // List of users associated with the location
	Devices []*DeviceWithComponents `json:"devices,omitempty"` // List of devices associated with the location
}

// LocationWithUsers represents a location with its associated users.
type LocationWithUsers struct {
	Location
	Users []*User `json:"users,omitempty"` // List of users associated with the location
}

// LocationWithDevices represents a location with its associated devices.
type LocationWithDevices struct {
	Location
	Devices []*DeviceWithComponents `json:"devices,omitempty"` // List of devices associated with the location
}
