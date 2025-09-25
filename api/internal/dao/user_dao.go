package dao

import (
	"CapIot-api/internal/models"
	"context"
)

// UserDAO defines the interface for user data access.
type UserDAO interface {
	CreateUser(ctx context.Context, auth0_id string, auth0_email string) (models.User, error)
	UserExists(ctx context.Context, auth0_id string) (models.User, error)
	FindUserByID(ctx context.Context, userID int) (*models.User, error)
	UpdateUser(ctx context.Context, user models.User) (*models.User, error)
	DeleteUser(ctx context.Context, userID int) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	AsignUser(ctx context.Context, userID, siteID int) error
	CountAll(ctx context.Context, search string) (int, error)
	UpdateUserName(ctx context.Context, id int, name string) error
	GetUserSites(ctx context.Context, id int) ([]models.Site, error)
	GetUsers(ctx context.Context, limit int, offset int, term string) ([]*models.User, error)
	UpdateUserSites(ctx context.Context, userID int, siteIDs []int) error
	GetUserLocations(ctx context.Context, userID int) ([]models.Location, error)
	GetUsernameByAuth0ID(ctx context.Context, id string) (models.User, error)
}
