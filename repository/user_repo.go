package repository

import (
	"api.cap.iot/config"
	"api.cap.iot/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetAllUsers retrieves all users
func GetAllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("users")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByAuth0ID finds a user by Auth0ID
func GetUserByAuth0ID(auth0ID string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"auth0Id": auth0ID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser inserts a new user with roles fetched from Auth0
func CreateUser(user *models.User, managementToken, auth0Domain string) error {
	// Fetch roles from Auth0 Management API
	roles, err := fetchRolesFromAuth0(user.Auth0ID, managementToken, auth0Domain)
	if err != nil {
		return err
	}
	user.Roles = roles

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("users")
	_, err = collection.InsertOne(ctx, user)
	return err
}

// UpdateUserRoles updates a user's roles in the database
func UpdateUserRoles(auth0ID, managementToken, auth0Domain string) error {
	roles, err := fetchRolesFromAuth0(auth0ID, managementToken, auth0Domain)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := config.GetCollection("users")
	filter := bson.M{"auth0Id": auth0ID}
	update := bson.M{"$set": bson.M{"roles": roles}}
	opts := options.Update().SetUpsert(true) // Create if not exists
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// fetchRolesFromAuth0 retrieves roles from Auth0 Management API
func fetchRolesFromAuth0(auth0ID, managementToken, auth0Domain string) ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	url := auth0Domain + "/api/v2/users/" + auth0ID + "/roles"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+managementToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch roles: %v", resp.Status)
	}

	var roles []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&roles); err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	return roleNames, nil
}
