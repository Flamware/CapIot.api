package handlers

import (
	"api.cap.iot/models"
	"api.cap.iot/repository"
	"encoding/json"
	"log"
	"net/http"
)

// CallbackHandler handles the user profile logic after Auth0 callback
func CallbackHandler(w http.ResponseWriter, r *http.Request, managementToken string, auth0Domain string) {
	// Extract user info from JWT claims (assuming the claims are set correctly in the middleware)
	claims, ok := r.Context().Value("jwtClaims").(map[string]interface{})
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract the necessary information from JWT claims
	auth0ID, _ := claims["sub"].(string)

	if auth0ID == "" {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if the user already exists in MongoDB
	existingUser, err := repository.GetUserByAuth0ID(auth0ID)
	if err == nil && existingUser != nil {
		// If the user already exists, return the user data
		json.NewEncoder(w).Encode(existingUser)
		return
	}

	// If user doesn't exist, create a new user
	newUser := models.NewUser(auth0ID)
	err = repository.CreateUser(newUser, managementToken, auth0Domain)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		log.Println("Error creating user:", err)
		return
	}

	// Return the newly created user
	json.NewEncoder(w).Encode(newUser)
}
