package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"CapIot-api/internal/repository"
	"context"
	"fmt"
)

// UserService defines the interface for user-related business logic
type UserService interface {
	CreateUser(ctx context.Context, auth0ID, auth0Email string) (error, models.User)
	GetUserByID(ctx context.Context, id int) (*models.User, error)
	UpdateUser(ctx context.Context, user models.User) (*models.User, error)
	DeleteUser(ctx context.Context, id int) error
	AsignUser(ctx context.Context, userID, siteID int) error
	GetUserSites(ctx context.Context, id int) ([]models.Site, error)
	UpdateUserSites(ctx context.Context, userID int, siteIDs []int, newName string) error
	GetUsers(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error)
	GetUserLocations(ctx context.Context, userID int) ([]models.Location, error)
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
func (s *DefaultUserService) CreateUser(ctx context.Context, auth0ID, auth0Email string) (error, models.User) {
	// The service layer can define its own timeout if it needs to, but we'll pass the context down.
	existingUser, err := s.userDAO.UserExists(ctx, auth0ID)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err), models.User{}
	}
	if existingUser.ID != 0 {
		// User already exists, return existing user
		return nil, existingUser
	}

	// User does not exist, create a new one
	createdUser, err := s.userDAO.CreateUser(ctx, auth0ID, auth0Email)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err), models.User{}
	}
	return nil, createdUser
}

// GetUserByID retrieves a user by their internal ID
func (s *DefaultUserService) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userDAO.FindUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID %d: %w", id, err)
	}
	return user, nil
}

// GetUsers retrieves users with pagination and optional search term
func (s *DefaultUserService) GetUsers(ctx context.Context, page int, limit int, term string) (map[string]interface{}, error) {
	users, err := s.userDAO.GetUsers(ctx, limit, (page-1)*limit, term)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with pagination: %w", err)
	}

	totalUsers, err := s.userDAO.CountAll(ctx, term)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
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

// UpdateUser updates an existing user's information
func (s *DefaultUserService) UpdateUser(ctx context.Context, user models.User) (*models.User, error) {
	updatedUser, err := s.userDAO.UpdateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user with ID %d: %w", user.ID, err)
	}
	return updatedUser, nil
}

// DeleteUser removes a user by their internal ID
func (s *DefaultUserService) DeleteUser(ctx context.Context, id int) error {
	if err := s.userDAO.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user with ID %d: %w", id, err)
	}
	return nil
}

// AsignUser assigns a user to a site
func (s *DefaultUserService) AsignUser(ctx context.Context, userID, siteID int) error {
	if err := s.userDAO.AsignUser(ctx, userID, siteID); err != nil {
		return fmt.Errorf("failed to assign user with ID %d to site with ID %d: %w", userID, siteID, err)
	}
	return nil
}

// GetUserLocations retrieves all locations assigned to a user
func (s *DefaultUserService) GetUserLocations(ctx context.Context, userID int) ([]models.Location, error) {
	locations, err := s.userDAO.GetUserLocations(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations for user with ID %d: %w", userID, err)
	}
	return locations, nil
}

// UpdateUserSites updates the user's name and assigned sites.
func (s *DefaultUserService) UpdateUserSites(ctx context.Context, userID int, siteIDs []int, newName string) error {
	// Use a single context for the transaction
	if err := s.userDAO.UpdateUserSites(ctx, userID, siteIDs); err != nil {
		return fmt.Errorf("failed to update user sites: %w", err)
	}

	if err := s.userDAO.UpdateUserName(ctx, userID, newName); err != nil {
		return fmt.Errorf("failed to update user name: %w", err)
	}

	return nil
}

// GetUserSites retrieves all sites associated with a user
func (s *DefaultUserService) GetUserSites(ctx context.Context, id int) ([]models.Site, error) {
	sites, err := s.userDAO.GetUserSites(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sites for user with ID %d: %w", id, err)
	}
	return sites, nil
}
