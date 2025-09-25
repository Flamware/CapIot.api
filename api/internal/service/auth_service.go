package service

import (
	"CapIot-api/internal/dao"
	"CapIot-api/internal/repository"
	"CapIot-api/internal/utils"
	"context"
	"fmt"
	"log"
)

// AuthService interface (define this in internal/service/auth_service.go)
type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string) error
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
}

// DefaultAuthService handles the business logic for user authentication
type DefaultAuthService struct {
	authRepo      *repository.AuthRepository
	userRepo      dao.UserDAO
	deviceRepo    dao.DeviceDAO
	componentRepo dao.ComponentDAO
}

// NewAuthService creates a new DefaultAuthService
func NewAuthService(authRepo *repository.AuthRepository, userRepo dao.UserDAO, deviceRepo *repository.PostgresDeviceDAO, componentRepo *repository.PostgresComponentDAO) *DefaultAuthService {
	return &DefaultAuthService{
		authRepo:      authRepo,
		userRepo:      userRepo,
		deviceRepo:    deviceRepo,
		componentRepo: componentRepo,
	}
}

// Login authenticates a user and generates a JWT token.
func (s *DefaultAuthService) Login(ctx context.Context, email, password string) (string, error) {
	// Step 1: Authenticate with Auth0 and get user info with roles
	authResult, err := s.authRepo.AuthenticateWithAuth0(email, password)
	if err != nil {
		return "", err
	}

	// Step 2: Check if user exists in the database
	user, err := s.userRepo.UserExists(ctx, authResult.Auth0ID)
	if err != nil {
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}
	// Check if the email from Auth0 matches the email in our database (if user exists),if not update it
	if user.ID != 0 && user.Email != authResult.Email {
		user.Email = authResult.Email
		updatedUser, err := s.userRepo.UpdateUser(ctx, user)
		if err != nil {
			return "", fmt.Errorf("failed to update user email: %w", err)
		}
		user = *updatedUser
		log.Printf("User email updated to: %s", authResult.Email)
	}
	log.Printf("User %s authenticated with roles: %v", authResult.Email, authResult.Role)

	// Check if the user does not exist in the database and create them
	if user.ID == 0 {
		newUser, err := s.userRepo.CreateUser(ctx, authResult.Auth0ID, authResult.Email)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}
		user = newUser
		log.Printf("New user created with ID: %d and email: %s", newUser.ID, authResult.Email)
	} else {
		log.Printf("User already exists with ID: %d and email: %s", user.ID, authResult.Email)
	}

	// Extract username safely
	username := ""
	if user.Name != nil {
		username = *user.Name
	} else if authResult.Username != "" {
		// fallback to username from Auth0 if available
		username = authResult.Username
	} else {
		// fallback to email if no username is set
		username = user.Email
	}

	// Log the successful login
	log.Printf("User %s logged in successfully with User ID: %d and username: %s", authResult.Email, user.ID, username)

	// Step 3: Generate the custom JWT token, including roles
	customJWT, err := utils.GenerateCustomJWT(authResult, username, user.ID)
	if err != nil {
		return "", err
	}

	return customJWT, nil
}

// Register handles user registration
func (s *DefaultAuthService) Register(ctx context.Context, email, password string) error {
	// Step 1: Register with Auth0
	auth0ID, err := s.authRepo.RegisterWithAuth0(email, password)
	if err != nil {
		return fmt.Errorf("failed to register with Auth0: %w", err)
	}

	// Step 2: Create a new user in the database
	_, err = s.userRepo.CreateUser(ctx, auth0ID, email)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Log the successful registration
	log.Printf("User registered with Auth0 ID: %s and email: %s", auth0ID, email)

	return nil
}

// GetUserRoles retrieves the roles of a user by their Auth0 User ID
func (s *DefaultAuthService) GetUserRoles(ctx context.Context, auth0UserID string) ([]string, error) {
	roles, err := s.authRepo.GetUserRolesByAuth0ID(auth0UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user %s: %w", auth0UserID, err)
	}
	return roles, nil
}
