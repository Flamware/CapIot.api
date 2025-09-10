package dao

import (
	"CapIot-api/internal/models"
	"context"
)

type LocationDAO interface {
	InsertLocation(location models.Location) error
	LocationExists(ID int) (bool, error)
	GetAllLocations(ctx context.Context, page int, limit int, term string) ([]*models.Location, error)
	GetComponentsByLocationID(id string) ([]models.Component, error)
	GetLocationByID(id string) (models.Location, error)
	GetDevicesByLocationID(id string) ([]models.Device, error)
	GetLocationsDevicesUsers(ctx context.Context, page int, limit int, term string) ([]*models.LocationWithUsersDevices, error)
	CountAll(ctx context.Context, term string) (int, error)
	DeleteLocation(ctx context.Context, id string) error
	ModifyLocation(ctx context.Context, location models.Location) error
	CreateSite(ctx context.Context, site models.Site) (*models.Site, error)
	GetSitesWithPagination(ctx context.Context, page int, limit int, term string) ([]*models.Site, error)
	CountSites(ctx context.Context, term string) (int, error)
	DeleteSite(ctx context.Context, id int64) error
	CheckUserAccessToLocation(id int, id2 int64) (bool, error)
	CheckUserAccessToSite(id int, id2 int64) (bool, error)
	GetLocationsBySiteIDs(ctx context.Context, siteIDs []string, page int, limit int, term string) ([]*models.Location, error)
}
