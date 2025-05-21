package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"CapIot-api/internal/repository"
	"context"
	"fmt"
	"log"
)

// UserService defines the interface for user-related business logic
type UserService interface {
	CreateUser(auth0ID, auth0Email string) (int, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(user models.User) (*models.User, error)
	DeleteUser(id int) error
	GetAllUsers() ([]models.User, error)
	AsignUser(userID, locationID int) error
	GetUserLocations(userID int) ([]models.Location, error)
	GetUsersLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error)
	UpdateUserAndLocation(UserID int, newLocationID []int, newName string) error
}

// DefaultUserService is the concrete implementation of the UserService interface
type DefaultUserService struct {
	userDAO  dao.UserDAO
	authRepo *repository.AuthRepository
}

// NewUserService creates a new DefaultUserService instance, injecting the UserDAO
func NewUserService(userDAO dao.UserDAO, authRepo *repository.AuthRepository) *DefaultUserService {
	return &DefaultUserService{
		userDAO:  userDAO,
		authRepo: authRepo,
	}
}

// CreateUser checks if a user with the given Auth0 ID exists, and creates one if not
func (s *DefaultUserService) CreateUser(auth0ID, auth0Email string) (int, error) {
	userID, err := s.userDAO.UserExists(auth0ID)
	if err != nil {
		return 0, fmt.Errorf("failed to check if user exists: %w", err)
	}

	if userID != 0 {
		return userID, nil
	}

	userID, err = s.userDAO.CreateUser(auth0ID, auth0Email)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return userID, nil
}

// GetUserByID retrieves a user by their internal ID
func (s *DefaultUserService) GetUserByID(id int) (*models.User, error) {
	user, err := s.userDAO.FindUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID %d: %w", id, err)
	}
	return user, nil
}

// UpdateUser updates an existing user's information
func (s *DefaultUserService) UpdateUser(user models.User) (*models.User, error) {
	updatedUser, err := s.userDAO.UpdateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user with ID %d: %w", user.ID, err)
	}
	return updatedUser, nil
}

// DeleteUser removes a user by their internal ID
func (s *DefaultUserService) DeleteUser(id int) error {
	if err := s.userDAO.DeleteUser(id); err != nil {
		return fmt.Errorf("failed to delete user with ID %d: %w", id, err)
	}
	return nil
}

// GetAllUsers retrieves all users
func (s *DefaultUserService) GetAllUsers() ([]models.User, error) {
	users, err := s.userDAO.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

// AsignUser assigns a user to a location
func (s *DefaultUserService) AsignUser(userID, locationID int) error {
	if err := s.userDAO.AsignUser(userID, locationID); err != nil {
		return fmt.Errorf("failed to assign user with ID %d to location with ID %d: %w", userID, locationID, err)
	}
	return nil
}

// GetUsersLocations retrieves users and their locations
func (s *DefaultUserService) GetUsersLocations(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error) {
	log.Printf("GetUsersLocations called with page: %d, limit: %d, search: '%s'", page, limit, term)

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch paginated and filtered data from the repository
	users, err := s.userDAO.FindAllWithLocations(ctx, limit, offset, term)
	if err != nil {
		log.Printf("Error fetching paginated and filtered data: %v", err)
		return nil, err
	}
	// for each user, append the role based on Auth0ID using GetUserRoles in authRepo
	for _, user := range users {
		roles, err := s.authRepo.GetUserRolesByAuth0ID(user.Auth0ID)
		if err != nil {
			log.Printf("Error fetching roles for user %s: %v", user.Auth0ID, err)
			return nil, err
		}
		user.Role = roles
		log.Printf("User %s has roles: %v", user.Auth0ID, roles)
	}

	// Fetch total count of items based on the search criteria
	totalUsers, err := s.userDAO.CountAll(ctx, term) // Update CountAll to accept search
	if err != nil {
		log.Printf("Error fetching total count with search: %v", err)
		return nil, err
	}

	totalPages := (totalUsers + limit - 1) / limit

	response := map[string]interface{}{
		"data":        users,
		"currentPage": page,
		"pageSize":    limit,
		"totalItems":  totalUsers,
		"totalPages":  totalPages,
	}

	return response, nil
}

// GetUserLocations retrieves all locations assigned to a user
func (s *DefaultUserService) GetUserLocations(userID int) ([]models.Location, error) {
	locations, err := s.userDAO.GetUserLocations(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations for user with ID %d: %w", userID, err)
	}
	return locations, nil
}

// updateUserAndLocation updates the user's location and name
func (s *DefaultUserService) UpdateUserAndLocation(UserID int, newLocationID []int, newName string) error {
	// Update the user's location
	if err := s.userDAO.UpdateUserLocation(UserID, newLocationID); err != nil {
		return fmt.Errorf("failed to update user location: %w", err)
	}

	// Update the user's name
	if err := s.userDAO.UpdateUserName(UserID, newName); err != nil {
		return fmt.Errorf("failed to update user name: %w", err)
	}

	return nil
}
