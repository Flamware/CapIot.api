package service

import (
	"api.cap.iot/dao"
	"api.cap.iot/repository"
	"api.cap.iot/utils"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
)

// AuthService handles the business logic for user authentication
type AuthService struct {
	authRepo *repository.AuthRepository
	userRepo dao.UserDAO
}

// NewAuthService creates a new AuthService
func NewAuthService(authRepo *repository.AuthRepository, userRepo dao.UserDAO) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

// Claims structure for managing JWT tokens
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(email, password string) (string, error) {
	// Step 1: Authenticate with Auth0
	auth0ID, err, auth0Email := s.authRepo.AuthenticateWithAuth0(email, password)
	if err != nil {
		return "", err
	}

	// Step 2: Check if user exists in the database
	exists, err := s.userRepo.UserExists(auth0ID)
	if err != nil {
		// If there is any error other than sql.ErrNoRows, return it
		return "", fmt.Errorf("failed to check user existence: %w", err)
	}

	// Check if the user does not exist in the database
	if !exists {
		log.Printf("User with email %s not found, creating new user.", email)

		// Create a new user with the Auth0 ID and email
		err := s.userRepo.CreateUser(auth0ID, auth0Email)
		if err != nil {
			return "", fmt.Errorf("failed to create user: %w", err)
		}

		// Log the creation of the user
		log.Printf("New user created with Auth0 ID: %s", auth0ID)
	} else {
		// Log if user is found in the database
		log.Printf("User found in the database: %s", auth0ID)
	}

	// Step 3: Generate the custom JWT token
	customJWT, err := utils.GenerateCustomJWT(auth0Email)
	if err != nil {
		return "", err
	}

	return customJWT, nil
}
