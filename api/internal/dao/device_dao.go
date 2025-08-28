package dao

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
)

// DeviceDAO defines the interface for device data access.
type DeviceDAO interface {
	CreateDevice(*sql.Tx, *models.Device) error
	DeviceExists(deviceID string) (bool, error)
	GetAllDevices() ([]*models.Device, error)
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error
	GetcomponentsByDeviceID(id string) ([]*models.Component, error)
	GetComponentByID(ComponentID string) (*models.Component, error)
	CreateComponent(*sql.Tx, *models.Component) (*models.Component, error)
	LinkComponentToDevice(tx *sql.Tx, deviceID string, componentID string) error
	FindAllWithcomponentsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithComponentsAndLocation, error)
	CountAll(ctx context.Context, search string) (int, error)
}
