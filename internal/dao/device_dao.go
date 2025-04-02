package dao

import "CapIot-api/internal/models"

// DeviceDAO defines the interface for device data access.
type DeviceDAO interface {
	InsertDevice(device models.Device) error
	DeviceExists(deviceID string) (bool, error)
	GetAllDevices() ([]models.Device, error)
	SetDeviceToLocation(deviceID string, locationID int) error
	IsDeviceAssigned(deviceID string) (bool, error) // Add this method

}
