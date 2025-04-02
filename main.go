package main

import (
	"api.cap.iot/config"
	"api.cap.iot/repository"
	"api.cap.iot/route"
	"api.cap.iot/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	userRepo := repository.NewPostgresUserRepository(db)
	deviceRepo := repository.NewPostgresDeviceRepository(db)
	locationRepo := repository.NewPostgresLocationRepository(db)

	authRepo, err := repository.NewAuthRepository()
	if err != nil {
		log.Fatalf("❌ Failed to initialize AuthRepository: %v", err)
	}

	// Initialize services
	authService := service.NewAuthService(authRepo, userRepo)
	deviceService := service.NewDeviceService(deviceRepo)
	locationService := service.NewLocationService(locationRepo)
	userService := service.NewUserService(userRepo)

	// MQTT client.
	mqttBroker := "tcp://localhost:1883"
	opts := mqtt.NewClientOptions().
		AddBroker(mqttBroker).
		SetCleanSession(true).
		SetUsername("admin").
		SetPassword("admin")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	log.Println("Successfully connected to MQTT broker!")

	// Set up the router with the service
	mux := route.SetupRouter(
		authService,
		deviceService,
		locationService,
		userService,
		client)

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
