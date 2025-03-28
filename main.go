package main

import (
	"api.cap.iot/config"
	"api.cap.iot/middleware"
	"api.cap.iot/models"
	"api.cap.iot/repository"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Connect to MongoDB
	log.Println("🔌 Initializing MongoDB connection...")
	config.ConnectDB()

	// Auth0 Configuration
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	auth0Audience := os.Getenv("AUTH0_AUDIENCE")
	managementToken := os.Getenv("AUTH0_MANAGEMENT_TOKEN")

	if auth0Domain == "" || auth0Audience == "" || managementToken == "" {
		log.Fatal("❌ AUTH0_DOMAIN, AUTH0_AUDIENCE, or AUTH0_MANAGEMENT_TOKEN is missing")
	}

	// Initialize routes
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/api/public", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "This is a public endpoint"})
	})

	// Protected callback endpoint with JWT validation and user creation
	mux.Handle("/api/callback", middleware.EnsureValidToken()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callbackHandler(w, r, managementToken, auth0Domain)
	})))

	// Admin-only endpoint
	mux.Handle("/api/admin", middleware.EnsureValidToken()(middleware.RequireRole("admin", managementToken, auth0Domain)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Welcome, admin!"})
	}))))

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	// Start the server
	log.Println("🚀 Starting server on :8080...")
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

// callbackHandler handles requests and syncs user data
func callbackHandler(w http.ResponseWriter, r *http.Request, managementToken, auth0Domain string) {
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
