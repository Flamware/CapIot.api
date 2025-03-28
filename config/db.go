package config

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func ConnectDB() {
	// MongoDB connection string (ensure it's correct)
	mongoURI := "mongodb://root:rootpassword@127.0.0.1:27017/capIot?authSource=admin"

	log.Println("🔌 Connecting to MongoDB...")

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("❌ MongoDB connection failed: %v", err)
	}

	log.Println("🔄 Attempting to ping MongoDB...")

	// Ping MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("❌ Could not ping MongoDB: %v", err)
	}

	log.Println("✅ MongoDB Connected Successfully!")

	// Assign the connected client to the global variable
	Client = client

	if Client != nil {
		log.Println("🚀 MongoDB Client is set!")
	} else {
		log.Println("❌ MongoDB Client is still nil!")
	}
}

func GetCollection(collectionName string) *mongo.Collection {
	if Client == nil {
		log.Fatal("❌ MongoDB Client is not initialized. Did you call ConnectDB() first?")
	}
	return Client.Database("capIot").Collection(collectionName) // 👈 Use "cap.iot" database
}
