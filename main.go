package main

import (
	"api.cap.iot/config"
	"api.cap.iot/dao"
	"api.cap.iot/repository"
	"api.cap.iot/route"
	"api.cap.iot/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/cors"
	"log"
	"net/http"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Connect to the database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("❌ Failed to initialize database connection: %v", err)
	}

	// Initialize DAO and repositories
	userDAO := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDAO)

	authRepo, err := repository.NewAuthRepository()
	if err != nil {
		log.Fatalf("❌ Failed to initialize AuthRepository: %v", err)
	}

	// Initialize services
	authService := service.NewAuthService(authRepo, userRepo)

	// Set up the router with the service
	mux := route.SetupRouter(authService)

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap the mux with CORS handler
	handler := c.Handler(mux)

	// Start the server
	log.Println("🚀 Starting server on :8080...")
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
