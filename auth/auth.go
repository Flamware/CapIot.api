package auth

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"os"
)

// ValidateJWT validates a JWT token and returns a boolean indicating if the token is valid
func ValidateJWT(tokenString string) (bool, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	}

	return false, errors.New("invalid token")
}
