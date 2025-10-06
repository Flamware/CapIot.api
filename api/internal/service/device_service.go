package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
)

// DeviceService interface defines the business logic for devices.
// It combines operations from both DAOs at the service level.
type DeviceService interface {
	// Transaction management
	BeginTransaction() (*sql.Tx, error)

	// Device operations
	CreateDevice(tx *sql.Tx, device *models.Device) error
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error
	UpdateDeviceOperationalStatus(tx *sql.Tx, deviceID string, operationalStatus models.OperationalStatus) error
	GetDeviceByID(id string) (*models.Device, error)
	GetAllDevices() ([]*models.Device, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(tx *sql.Tx, id string) error
	UnassignDeviceFromLocation(tx *sql.Tx, id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
	GetDevicescomponentsLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error)
	SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error
	CheckDeviceAccess(idInt int, id string) (bool, error)

	// Component operations
	CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error)
	GetComponentByID(id string) (*models.Component, error)
	LinkComponentToDevice(tx *sql.Tx, deviceID string, ComponentID string) error
	UpdateDeviceComponentStatus(tx *sql.Tx, id string, status string) error
	GetComponentsByDeviceID(id string) ([]*models.Component, error)
	UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error
	GetSensorsByDeviceID(id string) ([]*models.Component, error)
	ResetComponentRunningHours(tx *sql.Tx, id string) error
	UpdateComponentConfig(tx *sql.Tx, config models.ComponentConfig) error

	// Log operations
	HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error
	GetcomponentLogsByDeviceIDAndComponentID(id string, id2 string) ([]*models.ComponentLog, error)
	GetDeviceLogsByDeviceID(id string) ([]*models.ComponentLog, error)
	GetcomponentLogsByComponentID(id string) ([]*models.ComponentLog, error)
	GetAllLogsByUser(userId int) ([]*models.ComponentLog, error)
	UserHasAccessToComponent(id int, id2 string) (bool, error)
	MarkComponentLogsAsRead(tx *sql.Tx, ComponentID string, logIds []int) error
	MarkAllLogsAsRead(tx *sql.Tx, userID int) error

	// Schedule operations
	CreateRecurringSchedule(ctx context.Context, schedule *models.RecurringSchedule) (*models.RecurringSchedule, error)
	GetRecurringScheduleByID(ctx context.Context, scheduleID int) (*models.RecurringSchedule, error)
	GetRecurringSchedulesByDevice(ctx context.Context, deviceID string) ([]*models.RecurringSchedule, error)
	UpdateRecurringSchedule(ctx context.Context, scheduleID int, schedule *models.RecurringSchedule) error
	DeleteRecurringSchedule(ctx context.Context, scheduleID int) error
	UpdateDeviceConsumption(tx *sql.Tx, id string, current *float64, voltage *float64, power *float64) error
	UpdateDeviceProvisioningToken(id string, token string) error
	CheckDeviceToken(token string, id string) (bool, error)
	CheckDeviceLocation(id string, id2 string) (bool, error)
}

// DefaultDeviceService implements the DeviceService interface
type DefaultDeviceService struct {
	deviceDAO    dao.DeviceDAO
	componentDAO dao.ComponentDAO
	scheduleDAO  dao.ScheduleDAO // <-- Add the new dependency
}

// NewDeviceService creates a new DefaultDeviceService instance
func NewDeviceService(deviceDAO dao.DeviceDAO, componentDAO dao.ComponentDAO, scheduleDAO dao.ScheduleDAO) *DefaultDeviceService {
	return &DefaultDeviceService{
		deviceDAO:    deviceDAO,
		componentDAO: componentDAO,
		scheduleDAO:  scheduleDAO,
	}
}

// --- Transaction Management ---
func (s *DefaultDeviceService) BeginTransaction() (*sql.Tx, error) {
	// We only need one transaction, so we'll use the deviceDAO to start it.
	return s.deviceDAO.BeginTransaction()
}

// --- Device Operations ---
func (s *DefaultDeviceService) CreateDevice(tx *sql.Tx, device *models.Device) error {
	return s.deviceDAO.CreateDevice(tx, device)
}

func (s *DefaultDeviceService) GetDeviceByDeviceID(deviceID string) (*models.Device, error) {
	return s.deviceDAO.GetDeviceByID(deviceID)
}

func (s *DefaultDeviceService) UpdateDeviceLastSeenAndStatus(tx *sql.Tx, deviceID string, status models.OperationalStatus) error {
	return s.deviceDAO.UpdateDeviceLastSeenAndStatus(tx, deviceID, status)
}

func (s *DefaultDeviceService) UpdateDeviceOperationalStatus(tx *sql.Tx, deviceID string, operationalStatus models.OperationalStatus) error {
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
	// Get device and location info from deviceDAO
	devicesWithLocation, err := s.deviceDAO.FindAllWithComponentsAndLocations(ctx, limit, offset, search)
	if err != nil {
		log.Printf("Error fetching paginated and filtered data from deviceDAO: %v", err)
		return nil, err
	}
	// Now, get the components for each device using the componentDAO
	for _, deviceInfo := range devicesWithLocation {
		components, err := s.componentDAO.GetComponentsByDeviceID(deviceInfo.DeviceWithComponents.Device.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to get components for device %s: %w", deviceInfo.DeviceWithComponents.Device.DeviceID, err)
		}
		deviceInfo.DeviceWithComponents.Components = components
	}
	totalDevices, err := s.deviceDAO.CountAll(ctx, search)
	if err != nil {
		log.Printf("Error fetching total count with search: %v", err)
		return nil, err
	}
	totalPages := (totalDevices + limit - 1) / limit
	response := map[string]interface{}{
		"data":        devicesWithLocation,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalDevices,
		"totalPages":  totalPages,
	}
	return response, nil
}

func (s *DefaultDeviceService) SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error {

	// Check if device exists
	existingDevice, err := s.deviceDAO.GetDeviceByID(deviceID)
	if err != nil {
		return fmt.Errorf("error retrieving device '%s': %w", deviceID, err)
	}
	if existingDevice == nil {
		return fmt.Errorf("device '%s' not found", deviceID)
	}

	// Set device to location
	log.Printf("Setting device '%s' to location '%d'\n", deviceID, locationID)
	err = s.deviceDAO.SetDeviceToLocation(ctx, deviceID, locationID)
	if err != nil {
		return fmt.Errorf("failed to set device '%s' to location '%d': %w", deviceID, locationID, err)
	}

	return nil
}

func (s *DefaultDeviceService) CheckDeviceAccess(idInt int, id string) (bool, error) {
	return s.deviceDAO.CheckDeviceAccess(idInt, id)
}

// --- Component Operations ---
func (s *DefaultDeviceService) CreateComponent(tx *sql.Tx, component *models.Component) (*models.Component, error) {
	log.Printf("Attempting to create component with ID: '%s', Name: '%s' within transaction.\n", component.ComponentID, component.ComponentName)
	createdComponent, err := s.componentDAO.CreateComponent(tx, component)
	if err != nil {
		log.Printf("Error creating component in DAO for ID '%s': %v\n", component.ComponentID, err)
		return nil, fmt.Errorf("error creating component in DAO: %w", err)
	}
	log.Printf("Component '%s' with ID '%s' created successfully within transaction.\n", createdComponent.ComponentName, createdComponent.ComponentID)
	return createdComponent, nil
}

func (s *DefaultDeviceService) GetComponentByID(id string) (*models.Component, error) {
	return s.componentDAO.GetComponentByID(id)
}

func (s *DefaultDeviceService) LinkComponentToDevice(tx *sql.Tx, deviceID string, componentID string) error {
	log.Printf("Attempting to link component '%s' to device '%s' within transaction.\n", componentID, deviceID)
	err := s.componentDAO.LinkComponentToDevice(tx, deviceID, componentID)
	if err != nil {
		return fmt.Errorf("failed to link component '%s' to device '%s': %w", componentID, deviceID, err)
	}
	log.Printf("Linked component '%s' to device '%s' successfully within transaction.\n", componentID, deviceID)
	return nil
}

func (s *DefaultDeviceService) UpdateDeviceComponentStatus(tx *sql.Tx, id string, status string) error {
	log.Printf("Updating component status for component ID: '%s' to status: '%s' within transaction.", id, status)
	if strings.TrimSpace(id) == "" || strings.TrimSpace(status) == "" {
		return fmt.Errorf("component ID and status cannot be empty")
	}
	err := s.componentDAO.UpdateComponentStatus(tx, id, status)
	if err != nil {
		return fmt.Errorf("error updating component status for ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) GetComponentsByDeviceID(id string) ([]*models.Component, error) {
	components, err := s.componentDAO.GetComponentsByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving components for device ID '%s': %w", id, err)
	}
	return components, nil
}

func (s *DefaultDeviceService) UpdateComponentRunningHours(tx *sql.Tx, id string, hours int32) error {
	if strings.TrimSpace(id) == "" || hours < 0 {
		return fmt.Errorf("component ID cannot be empty and hours must be non-negative")
	}
	err := s.componentDAO.UpdateComponentRunningHours(tx, id, hours)
	if err != nil {
		return fmt.Errorf("error updating running hours for component ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) GetSensorsByDeviceID(id string) ([]*models.Component, error) {
	components, err := s.componentDAO.GetSensorsByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sensors for device ID '%s': %w", id, err)
	}
	return components, nil
}

func (s *DefaultDeviceService) ResetComponentRunningHours(tx *sql.Tx, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("component ID cannot be empty")
	}
	err := s.componentDAO.ResetComponentRunningHours(tx, id)
	if err != nil {
		return fmt.Errorf("error resetting running hours for component ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) UpdateComponentConfig(tx *sql.Tx, config models.ComponentConfig) error {
	err := s.componentDAO.UpdateComponentConfig(tx, config)
	if err != nil {
		return fmt.Errorf("error updating config for component ID '%s': %w", config.ComponentID, err)
	}
	return nil
}

// --- Log Operations ---
func (s *DefaultDeviceService) HandleDeviceAlert(tx *sql.Tx, componentID string, message string) error {
	log.Printf("Handling alert for component '%s': '%s'", componentID, message)
	err := s.componentDAO.HandleDeviceAlert(tx, componentID, message)
	if err != nil {
		return fmt.Errorf("error handling alert for component '%s': %w", componentID, err)
	}
	log.Printf("Alert handled successfully for component '%s'", componentID)
	return nil
}

func (s *DefaultDeviceService) GetcomponentLogsByDeviceIDAndComponentID(deviceID string, componentID string) ([]*models.ComponentLog, error) {
	log.Printf("Fetching logs for Device ID: '%s', component ID: '%s'", deviceID, componentID)
	if strings.TrimSpace(deviceID) == "" || strings.TrimSpace(componentID) == "" {
		return nil, fmt.Errorf("device ID and component ID cannot be empty")
	}
	logs, err := s.componentDAO.GetcomponentLogsByDeviceIDAndComponentID(deviceID, componentID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Device ID '%s' and component ID '%s': %w", deviceID, componentID, err)
	}
	log.Printf("Retrieved %d logs for Device ID: '%s', component ID: '%s'", len(logs), deviceID, componentID)
	return logs, nil
}

func (s *DefaultDeviceService) GetDeviceLogsByDeviceID(deviceID string) ([]*models.ComponentLog, error) {
	log.Printf("Fetching logs for Device ID: '%s'", deviceID)
	if strings.TrimSpace(deviceID) == "" {
		return nil, fmt.Errorf("device ID cannot be empty")
	}
	logs, err := s.componentDAO.GetDeviceLogsByDeviceID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Device ID '%s': %w", deviceID, err)
	}
	log.Printf("Retrieved %d logs for Device ID: '%s'", len(logs), deviceID)
	return logs, nil
}

func (s *DefaultDeviceService) GetcomponentLogsByComponentID(componentID string) ([]*models.ComponentLog, error) {
	log.Printf("Fetching logs for component ID: '%s'", componentID)
	if strings.TrimSpace(componentID) == "" {
		return nil, fmt.Errorf("component ID cannot be empty")
	}
	logs, err := s.componentDAO.GetcomponentLogsByComponentID(componentID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for component ID '%s': %w", componentID, err)
	}
	log.Printf("Retrieved %d logs for component ID: '%s'", len(logs), componentID)
	return logs, nil
}

func (s *DefaultDeviceService) GetAllLogsByUser(userId int) ([]*models.ComponentLog, error) {
	log.Println("Fetching all logs for the user")
	logs, err := s.componentDAO.GetAllLogsByUser(userId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all logs: %w", err)
	}
	log.Printf("Retrieved %d logs for the user", len(logs))
	return logs, nil
}

func (s *DefaultDeviceService) UserHasAccessToComponent(userId int, componentID string) (bool, error) {
	log.Printf("Checking access for User ID: '%d' to component ID: '%s'", userId, componentID)
	if userId <= 0 || strings.TrimSpace(componentID) == "" {
		return false, fmt.Errorf("invalid user ID or component ID")
	}
	hasAccess, err := s.componentDAO.UserHasAccessToComponent(userId, componentID)
	if err != nil {
		return false, fmt.Errorf("error checking access for User ID '%d' to component ID '%s': %w", userId, componentID, err)
	}
	log.Printf("User ID: '%d' has access to component ID: '%s': %v", userId, componentID, hasAccess)
	return hasAccess, nil
}

func (s *DefaultDeviceService) MarkComponentLogsAsRead(tx *sql.Tx, componentID string, logIds []int) error {
	log.Printf("Marking logs as read for component ID: '%s'", componentID)
	if strings.TrimSpace(componentID) == "" {
		return fmt.Errorf("component ID cannot be empty")
	}
	err := s.componentDAO.MarkComponentLogsAsRead(tx, componentID, logIds)
	if err != nil {
		return fmt.Errorf("error marking logs as read for component ID '%s': %w", componentID, err)
	}
	log.Printf("Successfully marked logs as read for component ID: '%s'", componentID)
	return nil
}

func (s *DefaultDeviceService) MarkAllLogsAsRead(tx *sql.Tx, userId int) error {
	log.Printf("Marking all logs as read for User ID: '%d'", userId)
	if userId <= 0 {
		return fmt.Errorf("invalid user ID")
	}
	err := s.componentDAO.MarkAllLogsAsRead(tx, userId)
	if err != nil {
		return fmt.Errorf("error marking all logs as read for User ID '%d': %w", userId, err)
	}
	log.Printf("Successfully marked all logs as read for User ID: '%d'", userId)
	return nil
}

// File: service/device_service.go (new method implementations)

// --- Schedule Operations ---
func (s *DefaultDeviceService) CreateRecurringSchedule(ctx context.Context, schedule *models.RecurringSchedule) (*models.RecurringSchedule, error) {
	scheduleID, err := s.scheduleDAO.CreateRecurringSchedule(ctx, *schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring schedule: %w", err)
	}
	schedule.RecurringScheduleID = scheduleID
	return schedule, nil
}

func (s *DefaultDeviceService) GetRecurringScheduleByID(ctx context.Context, scheduleID int) (*models.RecurringSchedule, error) {
	schedule, err := s.scheduleDAO.GetRecurringScheduleByID(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring schedule by ID: %w", err)
	}
	return schedule, nil
}

func (s *DefaultDeviceService) GetRecurringSchedulesByDevice(ctx context.Context, deviceID string) ([]*models.RecurringSchedule, error) {
	schedules, err := s.scheduleDAO.GetRecurringSchedulesByDevice(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring schedules for device %s: %w", deviceID, err)
	}
	// Convert []models.RecurringSchedule to []*models.RecurringSchedule
	result := make([]*models.RecurringSchedule, len(schedules))
	for i := range schedules {
		result[i] = &schedules[i]
	}
	return result, nil
}

func (s *DefaultDeviceService) UpdateRecurringSchedule(ctx context.Context, scheduleID int, schedule *models.RecurringSchedule) error {
	// You might add business logic here before calling the DAO, e.g., validation.
	return s.scheduleDAO.UpdateRecurringSchedule(ctx, scheduleID, *schedule)
}

func (s *DefaultDeviceService) DeleteRecurringSchedule(ctx context.Context, scheduleID int) error {
	return s.scheduleDAO.DeleteRecurringSchedule(ctx, scheduleID)
}

func (s *DefaultDeviceService) UpdateDeviceConsumption(tx *sql.Tx, id string, current *float64, voltage *float64, power *float64) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("device ID cannot be empty")
	}
	err := s.deviceDAO.UpdateDeviceConsumption(tx, id, current, voltage, power)
	if err != nil {
		return fmt.Errorf("error updating comsumption for device ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) UpdateDeviceProvisioningToken(id string, token string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("device ID cannot be empty")
	}
	err := s.deviceDAO.UpdateDeviceProvisioningToken(id, token)
	if err != nil {
		return fmt.Errorf("error updating provisioning token for device ID '%s': %w", id, err)
	}
	return nil
}

func (s *DefaultDeviceService) CheckDeviceToken(token string, id string) (bool, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(token) == "" {
		return false, fmt.Errorf("device ID and token cannot be empty")
	}
	isValid, err := s.deviceDAO.CheckDeviceToken(token, id)
	if err != nil {
		return false, fmt.Errorf("error checking token for device ID '%s': %w", id, err)
	}
	return isValid, nil
}

func (s *DefaultDeviceService) CheckDeviceLocation(id string, id2 string) (bool, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(id2) == "" {
		return false, fmt.Errorf("device ID and location ID cannot be empty")
	}
	isValid, err := s.deviceDAO.CheckDeviceLocation(id, id2)
	if err != nil {
		return false, fmt.Errorf("error checking location for device ID '%s': %w", id, err)
	}
	return isValid, nil
}
