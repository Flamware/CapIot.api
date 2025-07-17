package utils

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
)

const UserClaimsContextKey = "user_claims"

// GetUserIDFromContext retrieves the user ID (ID field) from the context.
func GetUserIDFromContext(ctx context.Context) (int, error) {
	// Retrieve claims from the context
	claims, ok := ctx.Value(UserClaimsContextKey).(jwt.MapClaims)
	if !ok || claims == nil {
		return 0, errors.New("user claims not found in context")
	}

	// Extract the user ID (ID field)
	userID, ok := claims["id"].(float64) // JSON numbers are parsed as float64
	if !ok {
		return 0, errors.New("user ID not found in claims")
	}

	return int(userID), nil
}
