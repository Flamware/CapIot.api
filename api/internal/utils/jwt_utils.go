package utils

import (
	"CapIot-api/internal/models"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"os"
	"time"
)

// GenerateCustomJWT generates a custom JWT token for the user
func GenerateCustomJWT(Authresult *models.AuthResult, user_id int) (string, error) {
	// Retrieve the secret key from environment variables
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return "", fmt.Errorf("secret key not found in environment variables")
	}

	// Create the JWT claims
	claims := jwt.MapClaims{
		"email": Authresult.Email,
		"sub":   Authresult.Auth0ID,
		"id":    user_id,
		"role":  Authresult.Role,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	// Create the token using the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
