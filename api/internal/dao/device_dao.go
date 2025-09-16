package dao

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
)

// DeviceDAO defines the interface for all device and related data access operations.
type DeviceDAO interface {
	// Transaction Operations
	BeginTransaction() (*sql.Tx, error)

	// Device Operations
	CreateDevice(tx *sql.Tx, device *models.Device) error
	DeviceExists(deviceID string) (bool, error)
	GetDeviceByID(id string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error
	UpdateDeviceOperationalStatus(tx *sql.Tx, id string, status models.OperationalStatus) error
	GetAllDevices() ([]*models.Device, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(tx *sql.Tx, id string) error
	CheckDeviceAccess(idInt int, id string) (bool, error)

	// Location Operations
	GetLocationByDeviceID(id string) (*models.Location, error)
	UnassignDeviceFromLocation(tx *sql.Tx, id string) error
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error
	FindAllWithComponentsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithComponentsAndLocation, error)
	CountAll(ctx context.Context, search string) (int, error)
}
