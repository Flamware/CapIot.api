package utils

import (
	"CapIot-api/internal/models"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"os"
	"time"
)

// Claims is a struct that holds the JWT claims
type Claims struct {
	Email    string   `json:"email"`
	Sub      string   `json:"sub"`
	ID       int      `json:"id"`
	Role     []string `json:"role"`
	Username string   `json:"username"`
	jwt.StandardClaims
}

// GenerateCustomJWT generates a custom JWT token for the user using the Claims struct
func GenerateCustomJWT(Authresult *models.AuthResult, username string, user_id int) (string, error) {
	// Retrieve the secret key from environment variables
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return "", fmt.Errorf("secret key not found in environment variables")
	}
	log.Printf("username: %s, user_id: %d", username, user_id)
	// Create the JWT claims using the Claims struct
	claims := Claims{
		Email:    Authresult.Email,
		Sub:      Authresult.Auth0ID,
		ID:       user_id,
		Role:     Authresult.Role,
		Username: username,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		},
	}
	log.Println("Generating JWT token with claims:", claims)
	// Create the token using the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRandomToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	// Encode the random bytes into a URL-safe base64 string
	return base64.URLEncoding.EncodeToString(b)
}
