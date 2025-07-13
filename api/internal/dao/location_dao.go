package dao

import (
	"CapIot-api/internal/models"
	"context"
)

type LocationDAO interface {
	InsertLocation(location models.Location) error
	LocationExists(ID int) (bool, error)
	GetAllLocations(ctx context.Context, page int, limit int, term string) ([]*models.Location, error)
	GetsensorsByLocationID(id string) ([]models.Sensor, error)
	GetLocationByID(id string) (models.Location, error)
	GetDevicesByLocationID(id string) ([]models.Device, error)
	GetLocationsDevicesUsers(ctx context.Context, page int, limit int, term string) ([]*models.LocationWithUsersDevices, error)
	CountAll(ctx context.Context, term string) (int, error)
	DeleteLocation(ctx context.Context, id string) error
	ModifyLocation(ctx context.Context, location models.Location) error
}
