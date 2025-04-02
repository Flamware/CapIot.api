package dao

import "CapIot-api/internal/models"

type LocationDAO interface {
	InsertLocation(location models.Location) error
	LocationExists(ID int) (bool, error)
	GetAllLocations() ([]models.Location, error)
}
