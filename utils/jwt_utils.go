package utils

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

// GenerateCustomJWT generates a custom JWT token for the user
func GenerateCustomJWT(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	secretKey := []byte("YOUR_SECRET_KEY")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
