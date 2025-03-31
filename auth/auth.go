package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"os"
)

// ValidateJWT validates a JWT token and returns the claims if the token is valid
func ValidateJWT(tokenString string) (jwt.MapClaims, error) {
	// Define a secret key to validate the token (this should be stored securely)
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
