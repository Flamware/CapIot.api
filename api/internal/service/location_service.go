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
