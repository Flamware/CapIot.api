package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"context"
	"log"
)

type DefaultLocationService struct {
	dao dao.LocationDAO
}

type LocationService interface {
	CreateLocation(location models.Location) error
	GetAllLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error)
	GetCaptorsByLocationID(locationID string) ([]models.Captor, error)
	GetDevicesByLocationID(id string) ([]models.Device, error)
	GetLocationsDevicesUsers(context context.Context, page int, limit int, term string) (map[string]interface{}, error)
	DeleteLocation(ctx context.Context, id string) error
	ModifyLocation(ctx context.Context, location models.Location) error
}

func NewLocationService(dao dao.LocationDAO) *DefaultLocationService {
	return &DefaultLocationService{dao: dao}
}

func (s *DefaultLocationService) CreateLocation(location models.Location) error {
	if location.ID != nil {

	}
	exists, err := s.dao.LocationExists(*location.ID)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return s.dao.InsertLocation(location)
}

func (s *DefaultLocationService) GetAllLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error) {
	log.Printf("GetAllLocations called with page: %d, limit: %d, search: '%s'", page, limit, term)

	// Calculate offset

	locations, err := s.dao.GetAllLocations(ctx, page, limit, term)
	if err != nil {
		log.Printf("Error fetching paginated locations: %v", err)
		return nil, err
	}

	// Fetch total count of items based on the search criteria
	totalLocations, err := s.dao.CountAll(ctx, term) // Assuming you have a CountAll in your Location DAO
	if err != nil {
		log.Printf("Error fetching total count of locations with search: %v", err)
		return nil, err
	}

	totalPages := (totalLocations + limit - 1) / limit

	response := map[string]interface{}{
		"data":        locations,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalLocations,
		"totalPages":  totalPages,
	}

	return response, nil
}

// service to get captors of a location
func (s *DefaultLocationService) GetCaptorsByLocationID(locationID string) ([]models.Captor, error) {
	captors, err := s.dao.GetCaptorsByLocationID(locationID)
	if err != nil {
		return nil, err
	}

	return captors, nil
}

// service to get devices of a location
func (s *DefaultLocationService) GetDevicesByLocationID(id string) ([]models.Device, error) {
	devices, err := s.dao.GetDevicesByLocationID(id)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

// service to get locations with users and devices (paginated)
func (s *DefaultLocationService) GetLocationsDevicesUsers(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error) {
	log.Printf("GetLocationsDevicesUsers called with page: %d, limit: %d, search: '%s'", page, limit, term)

	// Calculate offset
	locationsWithDetails, err := s.dao.GetLocationsDevicesUsers(ctx, page, limit, term)
	if err != nil {
		log.Printf("Error fetching paginated locations with devices and users: %v", err)
		return nil, err
	}

	// Fetch total count of items based on the search criteria
	totalLocations, err := s.dao.CountAll(ctx, term) // Assuming you have a CountAll in your Location DAO
	if err != nil {
		log.Printf("Error fetching total count of locations with search: %v", err)
		return nil, err
	}

	totalPages := (totalLocations + limit - 1) / limit

	response := map[string]interface{}{
		"data":        locationsWithDetails,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalLocations,
		"totalPages":  totalPages,
	}

	return response, nil
}

// service to delete a location
func (s *DefaultLocationService) DeleteLocation(ctx context.Context, id string) error {
	err := s.dao.DeleteLocation(ctx, id)
	if err != nil {
		log.Printf("Error deleting location with ID %s: %v", id, err)
		return err
	}

	return nil
}

// service to modify a location
func (s *DefaultLocationService) ModifyLocation(ctx context.Context, location models.Location) error {
	err := s.dao.ModifyLocation(ctx, location)
	if err != nil {
		log.Printf("Error modifying location: %v", err)
		return err
	}
	return nil
}
