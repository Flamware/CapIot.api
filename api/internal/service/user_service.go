package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/models"
	"fmt"
)

// UserService defines the interface for user-related business logic
type UserService interface {
	CreateUser(auth0ID, auth0Email string) (int, error)
	GetUserByID(id int) (*models.User, error)
	UpdateUser(user models.User) (*models.User, error)
	DeleteUser(id int) error
	GetAllUsers() ([]models.User, error)
}

// DefaultUserService is the concrete implementation of the UserService interface
type DefaultUserService struct {
	userDAO dao.UserDAO
}

// NewUserService creates a new DefaultUserService instance, injecting the UserDAO
func NewUserService(userDAO dao.UserDAO) UserService {
	return &DefaultUserService{
		userDAO: userDAO,
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
