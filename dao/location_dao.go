package dao

import "api.cap.iot/models"

type LocationDAO interface {
	InsertLocation(location models.Location) error
	LocationExists(ID int) (bool, error)
	GetAllLocations() ([]models.Location, error)
}
