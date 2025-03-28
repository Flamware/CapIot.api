package models

import "time"

// User represents a user in the system
type User struct {
	Auth0ID   string    `bson:"auth0Id" json:"auth0Id"`
	Roles     []string  `bson:"roles" json:"roles"` // Add this field
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

// NewUser creates a new User instance with the provided Auth0ID, Email, and Name
func NewUser(auth0ID string) *User {
	now := time.Now()
	return &User{
		Auth0ID:   auth0ID, // Set from Auth0 'sub' claim
		CreatedAt: now,     // Set to current time
	}
}
