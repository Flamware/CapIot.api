package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`        // ID from the database
	Email     string    `json:"email"`     // Email of the user
	Name      string    `json:"Name"`      // Email of the user
	Password  string    `json:"password"`  // Password of the user (it should be hashed before storing)
	Role      string    `json:"roles"`     // List of roles assigned to the user
	CreatedAt time.Time `json:"createdAt"` // Timestamp for when the user was created
}

// NewUser creates a new User instance with the provided Auth0ID and Email
func NewUser(auth0ID, email, password string, role string) *User {
	now := time.Now()
	return &User{
		Email:     email,    // Set the email
		Password:  password, // Set the user's password
		Role:      role,     // Set the user's roles
		CreatedAt: now,      // Set the current timestamp for user creation
	}
}
