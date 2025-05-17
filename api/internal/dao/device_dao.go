package dao

import (
	"CapIot-api/internal/models"
	"context"
)

// DeviceDAO defines the interface for device data access.
type DeviceDAO interface {
	InsertDevice(device *models.Device) error
	DeviceExists(deviceID string) (bool, error)
	GetAllDevices() ([]*models.Device, error)
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error
	IsDeviceAssigned(deviceID string) (bool, error) // Add this method
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)
	GetCaptorByID(captorID string) (*models.Captor, error)
	CreateCaptor(captor *models.Captor) (*models.Captor, error)
	InsertDeviceCaptor(deviceCaptor *models.DeviceCaptor) error
	FindAllWithSensorsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithSensorsAndLocation, error)
	CountAll(ctx context.Context, search string) (int, error)
}
