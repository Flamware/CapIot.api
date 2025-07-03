package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`        // ID from the database
	Auth0ID   string    `json:"auth0_id"`  // Auth0 ID of the user
	Email     string    `json:"email"`     // Email of the user
	Name      *string   `json:"name"`      // Email of the user
	Role      []string  `json:"roles"`     // List of roles assigned to the user
	CreatedAt time.Time `json:"createdAt"` // Timestamp for when the user was created
}

// NewUser creates a new User instance with the provided and Email
func NewUser(auth0ID, email, password string, role []string) *User {
	now := time.Now()
	return &User{
		Email:     email, // Set the email
		Role:      role,  // Set the role
		CreatedAt: now,   // Set the current timestamp for user creation
	}
}

// UsersLocations
type UserLocations struct {
	User
	Locations []*Location `json:"locations,omitempty"` // List of locations associated with the user
}
