package main

import (
	"CapIot-api/internal/config"
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/repository"
	"CapIot-api/internal/route"
	"CapIot-api/internal/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
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

	authRepo, err := repository.NewAuthRepository()
	if err != nil {
		log.Fatalf("❌ Failed to initialize AuthRepository: %v", err)
	}

	userRepo := repository.NewPostgresUserRepository(db)
	deviceRepo := repository.NewPostgresDeviceDAO(db)
	locationRepo := repository.NewPostgresLocationRepository(db)
	notificationRepo := repository.NewPostgresNotificationDAO(db)

	// MQTT client setup.
	mqttBroker := os.Getenv("MQTT_BROKER")
	log.Println("MQTT_BROKER:" + mqttBroker)
	if mqttBroker == "" {
		mqttBroker = "tcp://mqtt:1883" // Default to local Mosquitto
	}
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

	// Initialize services in two phases to resolve circular dependency:
	// 1. Initialize DeviceService with a nil MqttConfigPublisher initially
	//    We'll set it later after mqttHandler is created.
	deviceService := service.NewDeviceService(deviceRepo) // Pass nil for mqttPublisher initially

	// 2. Initialize MqttHandler, passing the (partially initialized) deviceService
	mqttHandler := handlers.NewMqttHandler(deviceService, client) // Pass deviceService here

	// 3. Now, set the MqttConfigPublisher on the deviceService to the mqttHandler.
	//    This completes the circular dependency injection.

	// Initialize other services (they don't have circular dependencies with MQTT handler)
	authService := service.NewAuthService(authRepo, userRepo, deviceRepo)
	locationService := service.NewLocationService(locationRepo)
	userService := service.NewUserService(userRepo, authRepo)
	notificationService := service.NewNotificationService(notificationRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	deviceHandler := handlers.NewDeviceHandler(deviceService)
	locationHandler := handlers.NewLocationHandler(locationService)
	userHandler := handlers.NewUserHandler(userService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	adminHandler := handlers.NewAdminHandler(
		authService,
		userService,
		deviceService,
		locationService,
	)

	// Set up the router with the handlers and MQTT client
	r := route.SetupRouter(
		authHandler,
		deviceHandler,
		locationHandler,
		userHandler,
		mqttHandler,
		authService,
		adminHandler,
		notificationHandler,
		client,
	)

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "CONNECT", "TRACE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	// Wrap the router with CORS handler
	handler := c.Handler(r) // Use the gorilla/mux router 'r'

	// Start the server
	log.Println("🚀 Starting server on :8080...")
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
