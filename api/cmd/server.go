package main

import (
	"CapIot-api/internal/config"
	repository2 "CapIot-api/internal/repository"
	"CapIot-api/internal/route"
	service2 "CapIot-api/internal/service"
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

	userRepo := repository2.NewPostgresUserRepository(db)
	deviceRepo := repository2.NewPostgresDeviceRepository(db)
	locationRepo := repository2.NewPostgresLocationRepository(db)

	authRepo, err := repository2.NewAuthRepository()
	if err != nil {
		log.Fatalf("❌ Failed to initialize AuthRepository: %v", err)
	}

	// Initialize services
	authService := service2.NewAuthService(authRepo, userRepo)
	deviceService := service2.NewDeviceService(deviceRepo)
	locationService := service2.NewLocationService(locationRepo)
	userService := service2.NewUserService(userRepo)

	// MQTT client.
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

	// Set up the router with the service
	mux := route.SetupRouter(
		authService,
		deviceService,
		locationService,
		userService,
		client)

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                                                                   // Allow all origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "CONNECT", "TRACE"}, // Allow all methods
		AllowedHeaders:   []string{"*"},                                                                   // Allow all headers
		AllowCredentials: true,                                                                            // If you need to allow credentials (cookies, auth headers, etc.)
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
