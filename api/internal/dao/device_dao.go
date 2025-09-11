package dao

import (
	"CapIot-api/internal/models"
	"context"
	"database/sql"
)

// DeviceDAO defines the interface for all device and related data access operations.
type DeviceDAO interface {
	BeginTransaction() (*sql.Tx, error)
	// Device Operations
	CreateDevice(tx *sql.Tx, device *models.Device) error // Updated to take tx
	DeviceExists(deviceID string) (bool, error)
	GetDeviceByID(id string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error // Updated to take tx
	UpdateDeviceOperationalStatus(tx *sql.Tx, id string, status models.OperationalStatus) error       // Updated to take tx
	GetAllDevices() ([]*models.Device, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(tx *sql.Tx, id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
	UnassignDeviceFromLocation(tx *sql.Tx, id string) error // Updated to take tx
	FindAllWithComponentsAndLocations(ctx context.Context, limit int, offset int, search string) ([]*models.DeviceWithComponentsAndLocation, error)
	CountAll(ctx context.Context, search string) (int, error)
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error

	// Component Operations
	CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error) // Updated to take tx
	GetComponentByID(id string) (*models.Component, error)
	LinkComponentToDevice(tx *sql.Tx, deviceID string, componentID string) error // Updated to take tx
	UpdateComponentRange(tx *sql.Tx, component *models.Component) error          // Updated to take tx
	UpdateComponentStatus(tx *sql.Tx, id string, status string) error            // Updated to take tx
	GetcomponentsByDeviceID(id string) ([]*models.Component, error)

	// Log Operations
	HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error // Updated to take tx
	GetcomponentLogsByComponentID(id string) ([]*models.ComponentLog, error)
	GetcomponentLogsByDeviceIDAndComponentID(deviceID string, ComponentID string) ([]*models.ComponentLog, error)
	GetDeviceLogsByDeviceID(deviceID string) ([]*models.ComponentLog, error)
	GetAllLogsByUser(userID int) ([]*models.ComponentLog, error)
	UserHasAccessTocomponent(userID int, ComponentID string) (bool, error)
	MarkcomponentLogsAsRead(tx *sql.Tx, ComponentID string, logIds []int) error // Updated to take tx
	MarkAllLogsAsRead(tx *sql.Tx, userID int) error
	UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error
	CheckDeviceAccess(idInt int, id string) (bool, error)
	GetSensorsByDeviceID(id string) ([]*models.Component, error)
}
