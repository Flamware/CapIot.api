// internal/service/device_service.go

package service

import (
	"CapIot-api/internal/models"
	"CapIot-api/internal/repository"
	"errors"
	"fmt"
	"log"
	"strings"
)

// DeviceService interface defines the business logic for devices
type DeviceService interface {
	CreateDevice(device *models.Device) error
	GetDeviceByDeviceID(deviceID string) (*models.Device, error)
	UpdateDeviceLastSeenAndStatus(deviceID string) error
	CreateCaptor(captor *models.Captor) (*models.Captor, error)
	SetDeviceToLocation(deviceID string, locationID int) error
	GetAllDevices() ([]*models.Device, error) // Ensure this return type
	GetOrCreateCaptor(captorID, captorName, captorType string) (*models.Captor, error)
	GetCaptorByID(id string) (*models.Captor, error)
	LinkCaptorToDevice(id string, id2 string) error
	UpdateDeviceOperationalStatus(id string, status models.OperationalStatus) error
	GetDeviceByID(id string) (*models.Device, error)
	GetCaptorsByDeviceID(id string) ([]*models.Captor, error)
	GetUnassignedDevices() ([]*models.Device, error)
	DeleteDevice(id string) error
	UnassignDeviceFromLocation(id string) error
	GetLocationByDeviceID(id string) (*models.Location, error)
}

// DefaultDeviceService implements the DeviceService interface
type DefaultDeviceService struct {
	deviceDAO repository.DeviceDAO
}

// NewDeviceService creates a new DefaultDeviceService instance
func NewDeviceService(dao *repository.PostgresDeviceDAO) *DefaultDeviceService {
	return &DefaultDeviceService{deviceDAO: dao}
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

func (s *DefaultDeviceService) CreateCaptor(captor *models.Captor) (*models.Captor, error) {
	log.Printf("Attempting to create captor with ID: '%s', Type: '%s'\n", captor.CaptorID, captor.CaptorType)

	// Business logic: Check if a captor with the same ID already exists
	existingCaptor, err := s.deviceDAO.GetCaptorByID(captor.CaptorID)
	if err != nil {
		log.Printf("Error checking for existing captor with ID '%s': %v\n", captor.CaptorID, err)
	}
	if existingCaptor != nil {
		log.Printf("Captor with ID '%s' already exists.\n", captor.CaptorID)
		return existingCaptor, fmt.Errorf("captor with ID '%s' already exists", captor.CaptorID)
	}

	// Business logic: Sanitize or validate captor data
	if captor.CaptorID == "" {
		log.Println("Error: Captor ID cannot be empty.")
		return nil, errors.New("captor ID cannot be empty")
	}
	if captor.CaptorType == "" {
		log.Println("Error: Captor Type cannot be empty.")
		return nil, errors.New("captor name cannot be empty")
	}

	// Call the DAO to create the captor
	createdCaptor, err := s.deviceDAO.CreateCaptor(captor)
	if err != nil {
		log.Printf("Error creating captor in DAO for ID '%s': %v\n", captor.CaptorID, err)
		return nil, fmt.Errorf("error creating captor in DAO: %w", err)
	}

	// Business logic: Optionally perform actions after successful creation
	log.Printf("Captor '%s' with ID '%s' created successfully.\n", createdCaptor.CaptorType, createdCaptor.CaptorID)

	return createdCaptor, nil
}

// SetDeviceToLocation calls the DAO to set a device to a specific location
func (s *DefaultDeviceService) SetDeviceToLocation(deviceID string, locationID int) error {
	log.Printf("DeviceService: Attempting to assign device '%s' to location '%d'\n", deviceID, locationID)

	// 1. Basic Input Validation (Optional but Recommended)
	if deviceID == "" {
		log.Println("DeviceService: Error - DeviceID cannot be empty")
		return fmt.Errorf("device ID cannot be empty")
	}
	if locationID <= 0 {
		log.Printf("DeviceService: Error - Invalid LocationID: %d\n", locationID)
		return fmt.Errorf("invalid location ID: %d", locationID)
	}

	// 2. Call the DAO and Handle Errors
	err := s.deviceDAO.SetDeviceToLocation(deviceID, locationID)
	if err != nil {
		log.Printf("DeviceService: Error from DAO when assigning device '%s' to location '%d': %v\n", deviceID, locationID, err)

		// 3. Provide More Contextual Error Messages (Consider different DAO error types)
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("device '%s' not found or location '%d' does not exist", deviceID, locationID)
		}
		// Add more specific error checks based on your DAO's error types
		return fmt.Errorf("failed to assign device '%s' to location '%d': %w", deviceID, locationID, err)
	}

	log.Printf("DeviceService: Successfully requested assignment of device '%s' to location '%d'\n", deviceID, locationID)
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

func (s *DefaultDeviceService) GetOrCreateCaptor(captorID, captorName, captorType string) (*models.Captor, error) {
	// Try to get the captor by ID first
	existingCaptor, err := s.deviceDAO.GetCaptorByID(captorID)
	if err != nil {
		return nil, fmt.Errorf("error checking for captor with ID '%s': %w", captorID, err)
	}
	if existingCaptor != nil {
		return existingCaptor, nil // Captor found by ID
	}

	// Captor not found by ID, try by name
	existingCaptor, err = s.deviceDAO.GetCaptorByID(captorID) // Assuming you have a GetCaptorByName in your DAO
	if err != nil {
		return nil, fmt.Errorf("error checking for captor with name '%s': %w", captorName, err)
	}
	if existingCaptor != nil {
		return existingCaptor, nil // Captor found by name
	}

	// Captor not found, create a new one
	newCaptor := &models.Captor{
		CaptorID:   captorID,
		CaptorType: captorType,
	}
	createdCaptor, err := s.deviceDAO.CreateCaptor(newCaptor)
	if err != nil {
		return nil, fmt.Errorf("error creating captor '%s' with ID '%s': %w", captorName, captorID, err)
	}
	return createdCaptor, nil
}

// GetCaptorByID retrieves a captor record by its ID (string)
func (s *DefaultDeviceService) GetCaptorByID(id string) (*models.Captor, error) {
	captor, err := s.deviceDAO.GetCaptorByID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving captor with ID '%s': %w", id, err)
	}
	return captor, nil
}

// LinkCaptorToDevice in DefaultDeviceService
func (s *DefaultDeviceService) LinkCaptorToDevice(deviceID string, captorID string) error {
	// ... (device and captor existence checks) ...

	deviceCaptor := &models.DeviceCaptor{
		DeviceID: deviceID,
		CaptorID: captorID,
	}
	return s.deviceDAO.InsertDeviceCaptor(deviceCaptor) // Correct call with the struct
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

// GetCaptorsByDeviceID retrieves captors associated with a device ID
func (s *DefaultDeviceService) GetCaptorsByDeviceID(id string) ([]*models.Captor, error) {
	captors, err := s.deviceDAO.GetCaptorsByDeviceID(id)
	if err != nil {
		return nil, fmt.Errorf("error retrieving captors for device ID '%s': %w", id, err)
	}
	return captors, nil
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
func (s *DefaultDeviceService) DeleteDevice(id string) error {
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
