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
	GetComponentsByLocationID(locationID string) ([]models.Component, error)
	GetDevicesByLocationID(id string) ([]models.Device, error)
	GetLocationsDevicesUsers(context context.Context, page int, limit int, term string) (map[string]interface{}, error)
	DeleteLocation(ctx context.Context, id string) error
	ModifyLocation(ctx context.Context, location models.Location) error
	CreateSite(ctx context.Context, site models.Site) (*models.Site, error)
	GetSitesWithPagination(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error)
	DeleteSite(ctx context.Context, id int64) error
}

func NewLocationService(dao dao.LocationDAO) *DefaultLocationService {
	return &DefaultLocationService{dao: dao}
}

func (s *DefaultLocationService) CreateLocation(location models.Location) error {
	log.Printf("CreateLocation called with location: %+v", location)
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

// service to get components of a location
func (s *DefaultLocationService) GetComponentsByLocationID(locationID string) ([]models.Component, error) {
	components, err := s.dao.GetComponentsByLocationID(locationID)
	if err != nil {
		return nil, err
	}

	return components, nil
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

// CreateSite creates a new site in the database and returns the created object.
func (s *DefaultLocationService) CreateSite(ctx context.Context, site models.Site) (*models.Site, error) {
	log.Printf("CreateSite called with site: %+v", site)

	// The DAO method should now return the created site with its new ID.
	createdSite, err := s.dao.CreateSite(ctx, site)
	if err != nil {
		log.Printf("Error creating site: %v", err)
		return nil, err
	}

	log.Printf("Site created successfully with ID: %d", *createdSite.ID)
	return createdSite, nil
}

// DeleteSite deletes a site by its ID.
func (s *DefaultLocationService) DeleteSite(ctx context.Context, id int64) error {
	log.Printf("DeleteSite called with ID: %d", id)

	err := s.dao.DeleteSite(ctx, id)
	if err != nil {
		log.Printf("Error deleting site with ID %d: %v", id, err)
		return err
	}

	log.Printf("Site with ID %d deleted successfully", id)
	return nil
}

// GetSitesWithPagination retrieves sites with pagination and optional search term.
func (s *DefaultLocationService) GetSitesWithPagination(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error) {
	log.Printf("GetSitesWithPagination called with page: %d, limit: %d, search: '%s'", page, limit, term)

	sites, err := s.dao.GetSitesWithPagination(ctx, page, limit, term)
	if err != nil {
		log.Printf("Error fetching paginated sites: %v", err)
		return nil, err
	}

	// Fetch total count of items based on the search criteria
	totalSites, err := s.dao.CountSites(ctx, term) // Assuming you have a CountSites in your Location DAO
	if err != nil {
		log.Printf("Error fetching total count of sites with search: %v", err)
		return nil, err
	}

	totalPages := (totalSites + limit - 1) / limit

	response := map[string]interface{}{
		"data":        sites,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalSites,
		"totalPages":  totalPages,
	}

	return response, nil
}
