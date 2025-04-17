package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
)

type DefaultLocationService struct {
	dao dao.LocationDAO
}

type LocationService interface {
	CreateLocation(location models.Location) error
	GetAllLocations() ([]models.Location, error)
	GetCaptorsByLocationID(locationID string) ([]models.Captor, error)
	GetDevicesByLocationID(id string) ([]models.Device, error)
}

func NewLocationService(dao dao.LocationDAO) *DefaultLocationService {
	return &DefaultLocationService{dao: dao}
}

func (s *DefaultLocationService) CreateLocation(location models.Location) error {
	exists, err := s.dao.LocationExists(location.ID)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return s.dao.InsertLocation(location)
}

func (s *DefaultLocationService) GetAllLocations() ([]models.Location, error) {
	return s.dao.GetAllLocations()
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
