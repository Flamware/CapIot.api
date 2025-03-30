package service

import (
	"api.cap.iot/models"
	"api.cap.iot/repository"
	"api.cap.iot/utils"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
	"time"
)

// AuthService handles the business logic for user authentication
type AuthService struct {
	authRepo *repository.AuthRepository // Change this to a pointer
	userRepo *repository.UserRepository // No change needed
}

// NewAuthService creates a new AuthService
func NewAuthService(authRepo *repository.AuthRepository, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		userRepo: userRepo,
	}
}

// Structure pour gérer les tokens JWT
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Clé secrète pour signer nos JWT (à stocker dans une variable d'environnement)
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Générer un JWT custom
func GenerateJWT(userID, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Expire en 24h

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Création du token signé
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Vérifier un JWT
func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
func (s *AuthService) Login(email, password string) (string, error) {
	// Step 1: Authenticate with Auth0
	// Authenticate with Auth0 but we won't use Auth0ID
	_, err := s.authRepo.AuthenticateWithAuth0(email, password)
	if err != nil {
		return "", err
	}

	// Step 2: Check if user exists in the database
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If user does not exist, create a new user
			log.Printf("User with email %s not found, creating new user.", email)

			// Create a new user with only the email (no Auth0ID)
			newUser := &models.User{
				Email: email, // Only store email
				// Optionally, you can also add 'role' or 'password' here if needed
			}

			// Attempt to create the user in the database
			email, err := s.userRepo.CreateUser(newUser)
			if err != nil {
				return "", fmt.Errorf("failed to create user: %w", err)
			}

			// Log the creation of the user (email is the only info we're storing)
			log.Printf("New user created with email: %s", email)
		} else {
			return "", fmt.Errorf("failed to check user existence: %w", err)
		}
	} else {
		// Log if user is found in the database
		log.Printf("User found in the database: %+v", user)
	}

	// Step 3: Generate the custom JWT token
	customJWT, err := utils.GenerateCustomJWT(email)
	if err != nil {
		return "", err
	}

	return customJWT, nil
}
