package repository

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Location represents a location structure
type Location struct {
	Name string `json:"name"`
}

// GetAvailableLocations fetches available locations from the API gateway
func GetAvailableLocations() ([]Location, error) {
	// Get API gateway URL and other configurations from environment variables
	apiGatewayURL := os.Getenv("API_GATEWAY_URL")
	clientID := os.Getenv("API_GATWAY_CLIEND_ID")
	clientSecret := os.Getenv("API_GATEWAY_CLIENT_SECRET")
	audience := os.Getenv("API_GATEWAY_AUDIENCE")

	if apiGatewayURL == "" || clientID == "" || clientSecret == "" || audience == "" {
		return nil, fmt.Errorf("one or more required environment variables are not set")
	}

	// Create a new HTTP client with a timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Create a new request
	req, err := http.NewRequest("GET", apiGatewayURL+"/locations", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the necessary headers
	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Client-Secret", clientSecret)
	req.Header.Set("Audience", audience)

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response body
	var locations []Location
	if err := json.NewDecoder(resp.Body).Decode(&locations); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %w", err)
	}

	return locations, nil
}
