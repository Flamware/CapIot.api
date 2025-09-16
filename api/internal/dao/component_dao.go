package dao

import (
	"CapIot-api/internal/models"
	"database/sql"
)

// ComponentDAO defines the interface for all component and related data access operations.
type ComponentDAO interface {
	// Component Operations
	CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error)
	GetComponentByID(id string) (*models.Component, error)
	LinkComponentToDevice(tx *sql.Tx, deviceID string, componentID string) error
	UpdateComponentRange(tx *sql.Tx, component *models.Component) error
	UpdateComponentStatus(tx *sql.Tx, id string, status string) error
	GetComponentsByDeviceID(id string) ([]*models.Component, error)
	UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error
	GetSensorsByDeviceID(id string) ([]*models.Component, error)
	ResetComponentRunningHours(tx *sql.Tx, id string) error
	UpdateComponentConfig(tx *sql.Tx, config models.ComponentConfig) error
	UserHasAccessToComponent(userID int, componentID string) (bool, error)

	// Log Operations
	HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error
	GetcomponentLogsByComponentID(id string) ([]*models.ComponentLog, error)
	GetcomponentLogsByDeviceIDAndComponentID(deviceID string, componentID string) ([]*models.ComponentLog, error)
	GetDeviceLogsByDeviceID(deviceID string) ([]*models.ComponentLog, error)
	GetAllLogsByUser(userID int) ([]*models.ComponentLog, error)
	MarkComponentLogsAsRead(tx *sql.Tx, componentID string, logIds []int) error
	MarkAllLogsAsRead(tx *sql.Tx, userID int) error
}
