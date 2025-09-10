package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"CapIot-api/internal/repository"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// MqttConfigPublisher defines the interface for publishing component configurations via MQTT.
// This decouples the service from the concrete MqttHandler implementation.
type MqttConfigPublisher interface {
	SetDeviceConfig(deviceID string, ComponentID string, minThreshold float64, maxThreshold float64) error
}

// DeviceService interface defines the business logic for devices
type DeviceService interface {
	// Transaction management
	BeginTransaction() (*sql.Tx, error) // Method to start a transaction

	// Device operations
	CreateDevice(tx *sql.Tx, device *models.Device) error                                                                // Now takes a transaction
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)                                                         // Read-only, no tx
	UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error                    // Now takes a transaction
	UpdateDeviceOperationalStatus(tx *sql.Tx, deviceID string, operationalStatus models.OperationalStatus) error         // Now takes a transaction
	GetDeviceByID(id string) (*models.Device, error)                                                                     // Read-only, no tx
	GetAllDevices() ([]*models.Device, error)                                                                            // Read-only, no tx
	GetUnassignedDevices() ([]*models.Device, error)                                                                     // Read-only, no tx
	DeleteDevice(tx *sql.Tx, id string) error                                                                            // No tx here, but DAO method might take it
	UnassignDeviceFromLocation(tx *sql.Tx, id string) error                                                              // Updated to take tx
	GetLocationByDeviceID(id string) (*models.Location, error)                                                           // Read-only, no tx
	GetDevicescomponentsLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error) // Read-only, no tx
	SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error
	CheckDeviceAccess(idInt int, id string) (bool, error)

	// Component operations
	CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error)                         // Now takes a transaction
	GetComponentByID(id string) (*models.Component, error)                                                      // Read-only, no tx
	LinkcomponentToDevice(tx *sql.Tx, deviceID string, ComponentID string) error                                // Now takes a transaction
	UpdateDeviceComponentStatus(tx *sql.Tx, id string, status string) error                                     // Now takes a transaction
	GetcomponentsByDeviceID(id string) ([]*models.Component, error)                                             // Read-only, no tx
	UpdateComponentRange(deviceID string, ComponentID string, minThreshold float64, maxThreshold float64) error // This method will manage its own transaction or be part of a larger one

	// Log operations
	HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error                         // Updated to take tx
	GetcomponentLogsByDeviceIDAndComponentID(id string, id2 string) ([]*models.ComponentLog, error) // Read-only, no tx
	GetDeviceLogsByDeviceID(id string) ([]*models.ComponentLog, error)                              // Read-only, no tx
	GetcomponentLogsByComponentID(id string) ([]*models.ComponentLog, error)                        // Read-only, no tx
	GetAllLogsByUser(userId int) ([]*models.ComponentLog, error)                                    // Read-only, no tx
	UserHasAccessTocomponent(id int, id2 string) (bool, error)                                      // Read-only, no tx
	MarkcomponentLogsAsRead(tx *sql.Tx, ComponentID string, logIds []int) error                     // Updated to take tx
	MarkAllLogsAsRead(tx *sql.Tx, userID int) error
	UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error
}

// DefaultDeviceService implements the DeviceService interface
type DefaultDeviceService struct {
	deviceDAO           dao.DeviceDAO
	MqttConfigPublisher MqttConfigPublisher // Inject the MQTT publisher here
}

// NewDeviceService creates a new DefaultDeviceService instance
// It now accepts an MqttConfigPublisher interface
func NewDeviceService(dao *repository.PostgresDeviceDAO, mqttPublisher MqttConfigPublisher) *DefaultDeviceService {
	return &DefaultDeviceService{
		deviceDAO:           dao,
		MqttConfigPublisher: mqttPublisher,
	}
}

// --- Transaction Management ---
func (s *DefaultDeviceService) BeginTransaction() (*sql.Tx, error) {
	return s.deviceDAO.BeginTransaction() // Call the DAO's BeginTransaction
}

// --- Device Operations ---
func (s *DefaultDeviceService) CreateDevice(tx *sql.Tx, device *models.Device) error {
	// The business logic of checking for existence is now handled by the caller (e.g., MqttHandler)
	// This method simply calls the DAO to create the device within the provided transaction.
	return s.deviceDAO.CreateDevice(tx, device)
}

func (s *DefaultDeviceService) GetDeviceByDeviceID(deviceID string) (*models.Device, error) {
	return s.deviceDAO.GetDeviceByID(deviceID)
}

func (s *DefaultDeviceService) UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error {
	return s.deviceDAO.UpdateDeviceLastSeenAndStatus(tx, deviceID, status)
}

func (s *DefaultDeviceService) UpdateDeviceOperationalStatus(tx *sql.Tx, deviceID string, operationalStatus models.OperationalStatus) error {
	// Business logic can go here, but the actual update happens in DAO with transaction
	return s.deviceDAO.UpdateDeviceOperationalStatus(tx, deviceID, operationalStatus)
}

func (s *DefaultDeviceService) GetDeviceByID(id string) (*models.Device, error) {
	return s.deviceDAO.GetDeviceByID(id)
}

func (s *DefaultDeviceService) GetAllDevices() ([]*models.Device, error) {
	devices, err := s.deviceDAO.GetAllDevices()
	if err != nil {
		return nil, fmt.Errorf("error fetching devices: %w", err)
	}
	return devices, nil
}

func (s *DefaultDeviceService) GetUnassignedDevices() ([]*models.Device, error) {
	unassignedDevices, err := s.deviceDAO.GetUnassignedDevices()
	if err != nil {
		return nil, fmt.Errorf("error retrieving unassigned devices: %w", err)
	}
	return unassignedDevices, nil
}

func (s *DefaultDeviceService) DeleteDevice(ctx *sql.Tx, id string) error {
	existingDevice, err := s.deviceDAO.GetDeviceByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving device '%s': %w", id, err)
	}
	if existingDevice == nil {
		return fmt.Errorf("device '%s' not found", id)
	}
	return s.deviceDAO.DeleteDevice(ctx, id)
}

func (s *DefaultDeviceService) UnassignDeviceFromLocation(tx *sql.Tx, id string) error {
	// Similar to DeleteDevice, might need its own transaction logic
	existingDevice, err := s.deviceDAO.GetDeviceByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving device '%s': %w", id, err)
	}
	if existingDevice == nil {
		return fmt.Errorf("device '%s' not found", id)
	}
	return s.deviceDAO.UnassignDeviceFromLocation(tx, id)
}

func (s *DefaultDeviceService) GetLocationByDeviceID(id string) (*models.Location, error) {
	location, err := s.deviceDAO.GetLocationByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving location for device ID '%s': %w", id, err)
	}
	return location, nil
}

func (s *DefaultDeviceService) GetDevicescomponentsLocations(ctx context.Context, page int, limit int, search string) (map[string]interface{}, error) {
	log.Printf("GetDevicescomponentsLocations called with page: %d, limit: %d, search: '%s'", page, limit, search)
	offset := (page - 1) * limit
	devices, err := s.deviceDAO.FindAllWithComponentsAndLocations(ctx, limit, offset, search)
	if err != nil {
		log.Printf("Error fetching paginated and filtered data: %v", err)
		return nil, err
	}
	totalDevices, err := s.deviceDAO.CountAll(ctx, search)
	if err != nil {
		log.Printf("Error fetching total count with search: %v", err)
		return nil, err
	}
	totalPages := (totalDevices + limit - 1) / limit
	response := map[string]interface{}{
		"data":        devices,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalDevices,
		"totalPages":  totalPages,
	}
	return response, nil
}

// GetcomponentsByDeviceID
func (s *DefaultDeviceService) GetcomponentsByDeviceID(id string) ([]*models.Component, error) {
	components, err := s.deviceDAO.GetcomponentsByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving components for device ID '%s': %w", id, err)
	}
	return components, nil
}

func (s *DefaultDeviceService) SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error {
	tx, err := s.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Pass the transaction context to the DAO layer
	ctxWithTx := context.WithValue(ctx, "tx", tx)
	err = s.deviceDAO.SetDeviceToLocation(ctxWithTx, deviceID, locationID)
	if err != nil {
		return fmt.Errorf("failed to set device to location: %w", err)
	}
	return nil
}

// --- Component Operations ---
func (s *DefaultDeviceService) CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error) {
	log.Printf("Attempting to create component with ID: '%s', Name: '%s' within transaction.\n", component.ComponentID, component.ComponentName)
	createdComponent, err := s.deviceDAO.CreateComponent(tx, component) // Pass the transaction
	if err != nil {
		log.Printf("Error creating component in DAO for ID '%s': %v\n", component.ComponentID, err)
		return nil, fmt.Errorf("error creating component in DAO: %w", err)
	}
	log.Printf("Component '%s' with ID '%s' created successfully within transaction.\n", createdComponent.ComponentName, createdComponent.ComponentID)
	return createdComponent, nil
}

func (s *DefaultDeviceService) GetComponentByID(id string) (*models.Component, error) {
	return s.deviceDAO.GetComponentByID(id)
}

func (s *DefaultDeviceService) LinkcomponentToDevice(tx *sql.Tx, deviceID string, ComponentID string) error {
	// Logic to check for existence is now handled by the caller (e.g., MqttHandler)
	// This method simply calls the DAO to link the component within the provided transaction.
	log.Printf("Attempting to link component '%s' to device '%s' within transaction.\n", ComponentID, deviceID)
	err := s.deviceDAO.LinkComponentToDevice(tx, deviceID, ComponentID) // Pass the transaction
	if err != nil {
		return fmt.Errorf("failed to link component '%s' to device '%s': %w", ComponentID, deviceID, err)
	}
	log.Printf("Linked component '%s' to device '%s' successfully within transaction.\n", ComponentID, deviceID)
	return nil
}

func (s *DefaultDeviceService) UpdateDeviceComponentStatus(tx *sql.Tx, id string, status string) error {
	// This method now updates the component's status within the provided transaction.
	log.Printf("Updating component status for component ID: '%s' to status: '%s' within transaction.", id, status)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(status) == "" {
		return fmt.Errorf("component ID and status cannot be empty")
	}
	err := s.deviceDAO.UpdateComponentStatus(tx, id, status) // Pass the transaction
	if err != nil {
		return fmt.Errorf("error updating component status for ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) UpdateComponentRange(deviceID string, ComponentID string, minThreshold float64, maxThreshold float64) error {
	// This method needs to manage its own transaction if it's not part of a larger one.
	// For simplicity, we'll create a new transaction here.
	tx, err := s.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for UpdateComponentRange: %w", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		} else if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// Validate device existence (read-only, can be outside transaction or use tx.QueryRow)
	device, err := s.deviceDAO.GetDeviceByID(deviceID)
	if err != nil {
		return fmt.Errorf("error retrieving device with ID '%s': %w", deviceID, err)
	}
	if device == nil {
		return fmt.Errorf("device with ID '%s' not found", deviceID)
	}

	// Validate component existence (read-only, can be outside transaction or use tx.QueryRow)
	component, err := s.deviceDAO.GetComponentByID(ComponentID)
	if err != nil {
		return fmt.Errorf("error retrieving component with ID '%s': %w", ComponentID, err)
	}
	if component == nil {
		return fmt.Errorf("component with ID '%s' not found", ComponentID)
	}

	// Update component thresholds (within the transaction)
	component.MinThreshold = &minThreshold
	component.MaxThreshold = &maxThreshold
	err = s.deviceDAO.UpdateComponentRange(tx, component) // Pass the transaction
	if err != nil {
		return fmt.Errorf("error updating component range for ID '%s': %w", ComponentID, err)
	}

	// Publish updated configuration via MQTT (this is usually outside the DB transaction)
	err = s.MqttConfigPublisher.SetDeviceConfig(deviceID, component.ComponentID, minThreshold, maxThreshold)
	if err != nil {
		log.Printf("Error publishing component config via MQTT for component ID '%s': %v", ComponentID, err)
		return fmt.Errorf("failed to publish component configuration: %w", err)
	}

	log.Printf("Component range updated successfully for ID '%s' and config published", ComponentID)
	return nil
}

// --- Log Operations ---
func (s *DefaultDeviceService) HandleDeviceAlert(tx *sql.Tx, ComponentID string, message string) error {
	// This method might need to manage its own transaction.
	// For simplicity, assuming DAO handles its own transaction or is part of a larger one.
	log.Printf("Handling alert for component '%s': '%s' ", ComponentID, message)
	err := s.deviceDAO.HandleDeviceAlert(tx, ComponentID, message)
	if err != nil {
		return fmt.Errorf("error handling alert for component '%s': %w", ComponentID, err)
	}
	log.Printf("Alert handled successfully for component '%s'", ComponentID)
	return nil
}

func (s *DefaultDeviceService) GetcomponentLogsByDeviceIDAndComponentID(deviceID string, ComponentID string) ([]*models.ComponentLog, error) {
	log.Printf("Fetching logs for Device ID: '%s', component ID: '%s'", deviceID, ComponentID)
	if strings.TrimSpace(deviceID) == "" || strings.TrimSpace(ComponentID) == "" {
		return nil, fmt.Errorf("device ID and component ID cannot be empty")
	}
	logs, err := s.deviceDAO.GetcomponentLogsByDeviceIDAndComponentID(deviceID, ComponentID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Device ID '%s' and component ID '%s': %w", deviceID, ComponentID, err)
	}
	log.Printf("Retrieved %d logs for Device ID: '%s', component ID: '%s'", len(logs), deviceID, ComponentID)
	return logs, nil
}

func (s *DefaultDeviceService) GetDeviceLogsByDeviceID(deviceID string) ([]*models.ComponentLog, error) {
	log.Printf("Fetching logs for Device ID: '%s'", deviceID)
	if strings.TrimSpace(deviceID) == "" {
		return nil, fmt.Errorf("device ID cannot be empty")
	}
	logs, err := s.deviceDAO.GetDeviceLogsByDeviceID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Device ID '%s': %w", deviceID, err)
	}
	log.Printf("Retrieved %d logs for Device ID: '%s'", len(logs), deviceID)
	return logs, nil
}

func (s *DefaultDeviceService) GetcomponentLogsByComponentID(ComponentID string) ([]*models.ComponentLog, error) {
	log.Printf("Fetching logs for component ID: '%s'", ComponentID)
	if strings.TrimSpace(ComponentID) == "" {
		return nil, fmt.Errorf("component ID cannot be empty")
	}
	logs, err := s.deviceDAO.GetcomponentLogsByComponentID(ComponentID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for component ID '%s': %w", ComponentID, err)
	}
	log.Printf("Retrieved %d logs for component ID: '%s'", len(logs), ComponentID)
	return logs, nil
}

func (s *DefaultDeviceService) GetAllLogsByUser(userId int) ([]*models.ComponentLog, error) {
	log.Println("Fetching all logs for the user")
	logs, err := s.deviceDAO.GetAllLogsByUser(userId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all logs: %w", err)
	}
	log.Printf("Retrieved %d logs for the user", len(logs))
	return logs, nil
}

func (s *DefaultDeviceService) UserHasAccessTocomponent(userId int, ComponentID string) (bool, error) {
	log.Printf("Checking access for User ID: '%d' to component ID: '%s'", userId, ComponentID)
	if userId <= 0 || strings.TrimSpace(ComponentID) == "" {
		return false, fmt.Errorf("invalid user ID or component ID")
	}
	hasAccess, err := s.deviceDAO.UserHasAccessTocomponent(userId, ComponentID)
	if err != nil {
		return false, fmt.Errorf("error checking access for User ID '%d' to component ID '%s': %w", userId, ComponentID, err)
	}
	log.Printf("User ID: '%d' has access to component ID: '%s': %v", userId, ComponentID, hasAccess)
	return hasAccess, nil
}

func (s *DefaultDeviceService) MarkcomponentLogsAsRead(tx *sql.Tx, ComponentID string, logIds []int) error {
	log.Printf("Marking logs as read for component ID: '%s'", ComponentID)
	if strings.TrimSpace(ComponentID) == "" {
		return fmt.Errorf("component ID cannot be empty")
	}
	err := s.deviceDAO.MarkcomponentLogsAsRead(tx, ComponentID, logIds)
	if err != nil {
		return fmt.Errorf("error marking logs as read for component ID '%s': %w", ComponentID, err)
	}
	log.Printf("Successfully marked logs as read for component ID: '%s'", ComponentID)
	return nil
}

func (s *DefaultDeviceService) MarkAllLogsAsRead(tx *sql.Tx, userId int) error {
	log.Printf("Marking all logs as read for User ID: '%d'", userId)
	if userId <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	err := s.deviceDAO.MarkAllLogsAsRead(tx, userId)
	if err != nil {
		return fmt.Errorf("error marking all logs as read for User ID '%d': %w", userId, err)
	}
	log.Printf("Successfully marked all logs as read for User ID: '%d'", userId)
	return nil
}

func (s *DefaultDeviceService) UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error {
	if strings.TrimSpace(id) == "" || hours < 0 {
		return fmt.Errorf("component ID cannot be empty and hours must be non-negative")
	}
	err := s.deviceDAO.UpdateComponentRunningHours(tx, id, hours) // Pass the transaction
	if err != nil {
		return fmt.Errorf("error updating running hours for component ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) CheckDeviceAccess(idInt int, id string) (bool, error) {
	return s.deviceDAO.CheckDeviceAccess(idInt, id)
}
