package dao

import (
	"CapIot-api/internal/models"
)

// UserDAO defines the interface for user data access.
type UserDAO interface {
	CreateUser(auth0_id string, auth0_email string) (int, error)
	UserExists(auth0_id string) (int, error)
	FindUserByID(userID int) (*models.User, error)
	UpdateUser(user models.User) (*models.User, error)
	DeleteUser(userID int) error
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]models.User, error) // Update the return type to include error
}
