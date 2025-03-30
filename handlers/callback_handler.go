package handlers

import (
	"api.cap.iot/middleware"
	"api.cap.iot/models"
	"api.cap.iot/repository"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func CallbackHandler(w http.ResponseWriter, r *http.Request, managementToken, auth0Domain string) {
	// Extract Auth0 User ID from context
	auth0UserID, ok := r.Context().Value(middleware.Auth0UserIDKey{}).(string)
	if !ok {
		log.Println("Could not get Auth0 user ID from context")
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("Authenticated user ID: %s", auth0UserID)

	// Check if user exists in the database
	user, err := repository.GetUserByAuth0ID(auth0UserID)
	if err != nil {
		// If user does not exist, create a new user with roles
		newUser := models.NewUser(auth0UserID)
		err = repository.CreateUser(newUser, managementToken, auth0Domain)
		if err != nil {
			log.Printf("Failed to create user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		log.Println("New user created:", newUser)
		user = newUser
	}

	// Send response
	response := map[string]string{
		"message": "Hello from a private endpoint! You need to be authenticated to see this.",
		"roles":   strings.Join(user.Roles, ", "),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
