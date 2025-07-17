package service

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
)

// MqttConfigPublisher defines the interface for publishing sensor configurations via MQTT.
// This decouples the service from the concrete MqttHandler implementation.
type MqttConfigPublisher interface {
	SetDeviceConfig(deviceID string, sensorID string, minThreshold float64, maxThreshold float64) error
}

// DeviceService interface defines the business logic for devices
type DeviceService interface {
	CreateDevice(device *models.Device) error
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(deviceID string) error
	Createsensor(sensor *models.Sensor) (*models.Sensor, error)
	SetDeviceToLocation(ctx context.Context, id string, id2 int) error
	GetAllDevices() ([]*models.Device, error)
	GetOrCreatesensor(SensorID, sensorName, SensorType string) (*models.Sensor, error)
	GetsensorByID(id string) (*models.Sensor, error)
	LinksensorToDevice(id string, id2 string) error
	UpdateDeviceOperationalStatus(id string, status models.OperationalStatus) error
	GetDeviceByID(id string) (*models.Device, error)
	GetsensorsByDeviceID(id string) ([]*models.Sensor, error)
	GetUnassignedDevices() ([]*models.Device, error)
	UnassignDeviceFromLocation(id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
	GetDevicesSensorsLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error)
	DeleteDevice(ctx context.Context, id string) error
	UpdatesensorRange(deviceID string, sensorID string, minThreshold float64, maxThreshold float64) error
	HandleDeviceAlert(sensorId string, message string) error
	GetSensorLogsByDeviceIDAndSensorID(id string, id2 string) ([]*models.SensorLog, error)
	GetDeviceLogsByDeviceID(id string) ([]*models.SensorLog, error)
	GetSensorLogsBySensorID(id string) ([]*models.SensorLog, error)
	GetAllLogsByUser(userId int) ([]*models.SensorLog, error)
	UserHasAccessToSensor(id int, id2 string) (bool, error)
	MarkSensorLogsAsRead(id string, logIds []int) error
	MarkAllLogsAsRead(id int) error
}

// DefaultDeviceService implements the DeviceService interface
type DefaultDeviceService struct {
	deviceDAO           repository.DeviceDAO
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

// CreateDevice calls the DAO to create a new device
func (s *DefaultDeviceService) CreateDevice(device *models.Device) error {
	return s.deviceDAO.CreateDevice(device)
}

// GetDeviceByDeviceID calls the DAO to retrieve a device by its ID
func (s *DefaultDeviceService) GetDeviceByDeviceID(deviceID string) (*models.Device, error) {
	return s.deviceDAO.GetDeviceByDeviceID(deviceID)
}

// UpdateDeviceLastSeenAndStatus calls the DAO to update the device's last seen and status
func (s *DefaultDeviceService) UpdateDeviceLastSeenAndStatus(deviceID string) error {
	return s.deviceDAO.UpdateDeviceLastSeenAndStatus(deviceID)
}

func (s *DefaultDeviceService) Createsensor(sensor *models.Sensor) (*models.Sensor, error) {
	log.Printf("Attempting to create sensor with ID: '%s', Type: '%s'\n", sensor.SensorID, sensor.SensorType)

	// Business logic: Check if a sensor with the same ID already exists
	existingsensor, err := s.deviceDAO.GetsensorByID(sensor.SensorID)
	if err != nil {
		log.Printf("Error checking for existing sensor with ID '%s': %v\n", sensor.SensorID, err)
	}
	if existingsensor != nil {
		log.Printf("sensor with ID '%s' already exists.\n", sensor.SensorID)
		return existingsensor, fmt.Errorf("sensor with ID '%s' already exists", sensor.SensorID)
	}

	// Business logic: Sanitize or validate sensor data
	if sensor.SensorID == "" {
		log.Println("Error: sensor ID cannot be empty.")
		return nil, errors.New("sensor ID cannot be empty")
	}
	if sensor.SensorType == "" {
		log.Println("Error: sensor Type cannot be empty.")
		return nil, errors.New("sensor name cannot be empty")
	}

	// Call the DAO to create the sensor
	createdsensor, err := s.deviceDAO.Createsensor(sensor)
	if err != nil {
		log.Printf("Error creating sensor in DAO for ID '%s': %v\n", sensor.SensorID, err)
		return nil, fmt.Errorf("error creating sensor in DAO: %w", err)
	}

	// Business logic: Optionally perform actions after successful creation
	log.Printf("sensor '%s' with ID '%s' created successfully.\n", createdsensor.SensorType, createdsensor.SensorID)

	return createdsensor, nil
}

// SetDeviceToLocation calls the DAO to set a device to a specific location
func (s *DefaultDeviceService) SetDeviceToLocation(ctx context.Context, deviceID string, locationID int) error {
	// Validate input
	if strings.TrimSpace(deviceID) == "" {
		fmt.Println("Error in service: Device ID cannot be empty") // Basic logging
		return fmt.Errorf("device ID cannot be empty")
	}
	if locationID <= 0 {
		fmt.Printf("Error in service: Invalid location ID: %d\n", locationID) // Basic logging
		return fmt.Errorf("invalid location ID: %d", locationID)
	}

	// First, check if the device exists
	existingDevice, err := s.deviceDAO.GetDeviceByDeviceID(deviceID)
	if err != nil {
		fmt.Printf("Error in service checking device existence: %v\n", err) // Basic logging
		return fmt.Errorf("error checking device existence: %w", err)
	}
	if existingDevice == nil {
		fmt.Printf("Error in service: Device '%s' not found\n", deviceID) // Basic logging
		return fmt.Errorf("device '%s' not found", deviceID)
	}
	// Check if device is already assigned to a location
	existingLocation, err := s.deviceDAO.GetLocationByDeviceID(deviceID)
	if err != nil {
		fmt.Printf("Error in service checking device location: %v\n", err) // Basic logging
		return fmt.Errorf("error checking device location: %w", err)
	}
	if existingLocation != nil {
		fmt.Printf("Error in service: Device '%s' is already assigned to a location\n", deviceID) // Basic logging
		return fmt.Errorf("device '%s' is already assigned to a location", deviceID)
	}

	// Call the DAO to set the device to the location
	err = s.deviceDAO.SetDeviceToLocation(ctx, deviceID, locationID)
	if err != nil {
		// Log the error with added service-level context
		fmt.Printf("Error in service setting device '%s' to location '%d': %v\n", deviceID, locationID, err)
		// Potentially add more service-specific error context or logic here
		// For example, you might check the type of the underlying error and decide to retry or handle it differently.
		return fmt.Errorf("failed to set device '%s' to location '%d': %w", deviceID, locationID, err)
	}
	return nil
}

// internal/service/device_service.go
func (s *DefaultDeviceService) GetAllDevices() ([]*models.Device, error) { // Ensure this return type
	devices, err := s.deviceDAO.GetAllDevices()
	if err != nil {
		return nil, fmt.Errorf("error fetching devices: %w", err)
	}
	return devices, nil
}

// internal/service/device_service.go

func (s *DefaultDeviceService) GetOrCreatesensor(SensorID, sensorName, SensorType string) (*models.Sensor, error) {
	// Try to get the sensor by ID first
	existingsensor, err := s.deviceDAO.GetsensorByID(SensorID)
	if err != nil {
		return nil, fmt.Errorf("error checking for sensor with ID '%s': %w", SensorID, err)
	}
	if existingsensor != nil {
		return existingsensor, nil // sensor found by ID
	}

	// sensor not found by ID, try by name
	existingsensor, err = s.deviceDAO.GetsensorByID(SensorID) // Assuming you have a GetsensorByName in your DAO
	if err != nil {
		return nil, fmt.Errorf("error checking for sensor with name '%s': %w", sensorName, err)
	}
	if existingsensor != nil {
		return existingsensor, nil // sensor found by name
	}

	// sensor not found, create a new one
	newsensor := &models.Sensor{
		SensorID:   SensorID,
		SensorType: SensorType,
	}
	createdsensor, err := s.deviceDAO.Createsensor(newsensor)
	if err != nil {
		return nil, fmt.Errorf("error creating sensor '%s' with ID '%s': %w", sensorName, SensorID, err)
	}
	return createdsensor, nil
}

// GetsensorByID retrieves a sensor record by its ID (string)
func (s *DefaultDeviceService) GetsensorByID(id string) (*models.Sensor, error) {
	sensor, err := s.deviceDAO.GetsensorByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sensor with ID '%s': %w", id, err)
	}
	return sensor, nil
}

// LinksensorToDevice in DefaultDeviceService
func (s *DefaultDeviceService) LinksensorToDevice(deviceID string, SensorID string) error {
	// ... (device and sensor existence checks) ...

	devicesensor := &models.Devicesensor{
		DeviceID: deviceID,
		SensorID: SensorID,
	}
	return s.deviceDAO.InsertDevicesensor(devicesensor) // Correct call with the struct
}

// UpdateDeviceOperationalStatus updates the operational status of a device
func (s *DefaultDeviceService) UpdateDeviceOperationalStatus(deviceID string, operationalStatus models.OperationalStatus) error {
	// Add any business logic before updating the operational status
	existingDevice, err := s.deviceDAO.GetDeviceByDeviceID(deviceID)
	if err != nil {
		return fmt.Errorf("error retrieving device '%s': %w", deviceID, err)
	}
	if existingDevice == nil {
		return fmt.Errorf("device '%s' not found", deviceID)
	}

	// Perform any validation or business rules related to operational status changes
	// For example, you might want to log the status change

	return s.deviceDAO.UpdateDeviceOperationalStatus(deviceID, operationalStatus)
}

// GetDeviceByID retrieves a device by its ID
func (s *DefaultDeviceService) GetDeviceByID(id string) (*models.Device, error) {
	device, err := s.deviceDAO.GetDeviceByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving device with ID '%s': %w", id, err)
	}
	return device, nil
}

// GetsensorsByDeviceID retrieves sensors associated with a device ID
func (s *DefaultDeviceService) GetsensorsByDeviceID(id string) ([]*models.Sensor, error) {
	sensors, err := s.deviceDAO.GetsensorsByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving sensors for device ID '%s': %w", id, err)
	}
	return sensors, nil
}

// GetUnassignedDevices retrieves devices that are not assigned to any location
func (s *DefaultDeviceService) GetUnassignedDevices() ([]*models.Device, error) {
	unassignedDevices, err := s.deviceDAO.GetUnassignedDevices()
	if err != nil {
		return nil, fmt.Errorf("error retrieving unassigned devices: %w", err)
	}
	return unassignedDevices, nil
}

// DeleteDevice deletes a device by its ID
func (s *DefaultDeviceService) DeleteDevice(ctx context.Context, id string) error {
	// Perform any necessary checks before deletion
	existingDevice, err := s.deviceDAO.GetDeviceByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving device '%s': %w", id, err)
	}
	if existingDevice == nil {
		return fmt.Errorf("device '%s' not found", id)
	}

	// Call the DAO to delete the device
	err = s.deviceDAO.DeleteDevice(id)
	if err != nil {
		return fmt.Errorf("error deleting device '%s': %w", id, err)
	}

	return nil
}

// UnassignDeviceFromLocation unassigns a device from its current location
func (s *DefaultDeviceService) UnassignDeviceFromLocation(id string) error {
	// Perform any necessary checks before unassignment
	existingDevice, err := s.deviceDAO.GetDeviceByID(id)
	if err != nil {
		return fmt.Errorf("error retrieving device '%s': %w", id, err)
	}
	if existingDevice == nil {
		return fmt.Errorf("device '%s' not found", id)
	}

	// Call the DAO to unassign the device from its current location
	err = s.deviceDAO.UnassignDeviceFromLocation(id)
	if err != nil {
		return fmt.Errorf("error unassigning device '%s': %w", id, err)
	}

	return nil
}

// GetLocationByDeviceID retrieves the location associated with a device ID
func (s *DefaultDeviceService) GetLocationByDeviceID(id string) (*models.Location, error) {
	location, err := s.deviceDAO.GetLocationByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving location for device ID '%s': %w", id, err)
	}
	return location, nil
}

// GetDevicesSensorsLocations retrieves devices, sensors, and locations for a user
func (s *DefaultDeviceService) GetDevicesSensorsLocations(ctx context.Context, page int, limit int, search string) (map[string]interface{}, error) {
	log.Printf("GetDevicesSensorsLocations called with page: %d, limit: %d, search: '%s'", page, limit, search)

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch paginated and filtered data from the repository
	devices, err := s.deviceDAO.FindAllWithSensorsAndLocations(ctx, limit, offset, search)
	if err != nil {
		log.Printf("Error fetching paginated and filtered data: %v", err)
		return nil, err
	}

	// Fetch total count of items based on the search criteria
	totalDevices, err := s.deviceDAO.CountAll(ctx, search) // Update CountAll to accept search
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

// UpdatesensorRange updates the operational range of a sensor
func (s *DefaultDeviceService) UpdatesensorRange(deviceID string, sensorID string, minThreshold float64, maxThreshold float64) error {
	// Validate device existence
	device, err := s.deviceDAO.GetDeviceByID(deviceID)
	if err != nil {
		return fmt.Errorf("error retrieving device with ID '%s': %w", deviceID, err)
	}
	if device == nil {
		return fmt.Errorf("device with ID '%s' not found", deviceID)
	}

	// Validate sensor existence
	sensor, err := s.deviceDAO.GetsensorByID(sensorID)
	if err != nil {
		return fmt.Errorf("error retrieving sensor with ID '%s': %w", sensorID, err)
	}
	if sensor == nil {
		return fmt.Errorf("sensor with ID '%s' not found", sensorID)
	}

	// Update sensor thresholds
	sensor.MinThreshold = &minThreshold
	sensor.MaxThreshold = &maxThreshold
	err = s.deviceDAO.UpdatesensorRange(sensor)
	if err != nil {
		return fmt.Errorf("error updating sensor range for ID '%s': %w", sensorID, err)
	}

	// Publish updated configuration via MQTT
	err = s.MqttConfigPublisher.SetDeviceConfig(deviceID, sensor.SensorID, minThreshold, maxThreshold)
	if err != nil {
		log.Printf("Error publishing sensor config via MQTT for sensor ID '%s': %v", sensorID, err)
		return fmt.Errorf("failed to publish sensor configuration: %w", err)
	}

	log.Printf("Sensor range updated successfully for ID '%s' and config published", sensorID)
	return nil
}

// HandleDeviceAlert processes device alerts
func (s *DefaultDeviceService) HandleDeviceAlert(SensorID string, message string) error {
	// Log the alert handling
	log.Printf("Handling alert for with sensor '%s':'%s' ", SensorID, message)

	// Perform any necessary business logic here
	// For example, you might want to log the alert to a database or notify an admin

	// Call the DAO to handle the alert
	err := s.deviceDAO.HandleDeviceAlert(SensorID, message)
	if err != nil {
		return fmt.Errorf("error handling alert for sensor '%s': %w", SensorID, err)
	}

	log.Printf("Alert handled successfully for Sensor '%s'", SensorID)
	return nil
}

// GetSensorLogsByDeviceIDAndSensorID retrieves sensor logs for a specific device and sensor
func (s *DefaultDeviceService) GetSensorLogsByDeviceIDAndSensorID(deviceID string, sensorID string) ([]*models.SensorLog, error) {
	log.Printf("Fetching logs for Device ID: '%s', Sensor ID: '%s'", deviceID, sensorID)

	// Validate input
	if strings.TrimSpace(deviceID) == "" || strings.TrimSpace(sensorID) == "" {
		return nil, fmt.Errorf("device ID and sensor ID cannot be empty")
	}

	// Call the DAO to get the logs
	logs, err := s.deviceDAO.GetSensorLogsByDeviceIDAndSensorID(deviceID, sensorID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Device ID '%s' and Sensor ID '%s': %w", deviceID, sensorID, err)
	}

	log.Printf("Retrieved %d logs for Device ID: '%s', Sensor ID: '%s'", len(logs), deviceID, sensorID)
	return logs, nil
}

// GetDeviceLogsByDeviceID retrieves all logs for a specific device
func (s *DefaultDeviceService) GetDeviceLogsByDeviceID(deviceID string) ([]*models.SensorLog, error) {
	log.Printf("Fetching logs for Device ID: '%s'", deviceID)

	// Validate input
	if strings.TrimSpace(deviceID) == "" {
		return nil, fmt.Errorf("device ID cannot be empty")
	}

	// Call the DAO to get the logs
	logs, err := s.deviceDAO.GetDeviceLogsByDeviceID(deviceID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Device ID '%s': %w", deviceID, err)
	}

	log.Printf("Retrieved %d logs for Device ID: '%s'", len(logs), deviceID)
	return logs, nil
}

// GetSensorLogsBySensorID retrieves all logs for a specific sensor
func (s *DefaultDeviceService) GetSensorLogsBySensorID(sensorID string) ([]*models.SensorLog, error) {
	log.Printf("Fetching logs for Sensor ID: '%s'", sensorID)

	// Validate input
	if strings.TrimSpace(sensorID) == "" {
		return nil, fmt.Errorf("sensor ID cannot be empty")
	}

	// Call the DAO to get the logs
	logs, err := s.deviceDAO.GetSensorLogsBySensorID(sensorID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving logs for Sensor ID '%s': %w", sensorID, err)
	}

	log.Printf("Retrieved %d logs for Sensor ID: '%s'", len(logs), sensorID)
	return logs, nil
}

// GetAllLogsByUser retrieves all logs for the user
func (s *DefaultDeviceService) GetAllLogsByUser(userId int) ([]*models.SensorLog, error) {
	log.Println("Fetching all logs for the user")

	// Call the DAO to get all logs
	logs, err := s.deviceDAO.GetAllLogsByUser(userId)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all logs: %w", err)
	}

	log.Printf("Retrieved %d logs for the user", len(logs))
	return logs, nil
}

// UserHasAccessToSensor checks if a user has access to a specific sensor
func (s *DefaultDeviceService) UserHasAccessToSensor(userId int, sensorID string) (bool, error) {
	log.Printf("Checking access for User ID: '%d' to Sensor ID: '%s'", userId, sensorID)

	// Validate input
	if userId <= 0 || strings.TrimSpace(sensorID) == "" {
		return false, fmt.Errorf("invalid user ID or sensor ID")
	}

	// Call the DAO to check access
	hasAccess, err := s.deviceDAO.UserHasAccessToSensor(userId, sensorID)
	if err != nil {
		return false, fmt.Errorf("error checking access for User ID '%d' to Sensor ID '%s': %w", userId, sensorID, err)
	}

	log.Printf("User ID: '%d' has access to Sensor ID: '%s': %v", userId, sensorID, hasAccess)
	return hasAccess, nil
}

// MarkSensorLogsAsRead marks sensor logs as read for a specific sensor
func (s *DefaultDeviceService) MarkSensorLogsAsRead(sensorID string, logIds []int) error {
	log.Printf("Marking logs as read for Sensor ID: '%s'", sensorID)

	// Validate input
	if strings.TrimSpace(sensorID) == "" {
		return fmt.Errorf("sensor ID cannot be empty")
	}

	// Call the DAO to mark logs as read
	err := s.deviceDAO.MarkSensorLogsAsRead(sensorID, logIds)
	if err != nil {
		return fmt.Errorf("error marking logs as read for Sensor ID '%s': %w", sensorID, err)
	}

	log.Printf("Successfully marked logs as read for Sensor ID: '%s'", sensorID)
	return nil
}

// MarkAllLogsAsRead marks all logs as read for a specific user
func (s *DefaultDeviceService) MarkAllLogsAsRead(userId int) error {
	log.Printf("Marking all logs as read for User ID: '%d'", userId)

	// Validate input
	if userId <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	// Call the DAO to mark all logs as read
	err := s.deviceDAO.MarkAllLogsAsRead(userId)
	if err != nil {
		return fmt.Errorf("error marking all logs as read for User ID '%d': %w", userId, err)
	}

	log.Printf("Successfully marked all logs as read for User ID: '%d'", userId)
	return nil
}
