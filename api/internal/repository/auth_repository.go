package repository

import (
	"CapIot-api/internal/models"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// AuthRepository handles interactions with external authentication providers like Auth0
type AuthRepository struct {
	auth0Domain  string
	clientID     string
	clientSecret string
	audience     string
}

// NewAuthRepository creates a new AuthRepository
func NewAuthRepository() (*AuthRepository, error) {
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	audience := os.Getenv("AUTH0_AUDIENCE")

	if auth0Domain == "" || clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("❌ Auth0 credentials are missing in environment variables")
	}

	authRepo := &AuthRepository{
		auth0Domain:  auth0Domain,
		clientID:     clientID,
		clientSecret: clientSecret,
		audience:     audience,
	}

	fmt.Println("✅ AuthRepository initialized successfully") // Debug log

	return authRepo, nil
}

// AuthenticateWithAuth0 verifies the user's credentials using Auth0's API
func (r *AuthRepository) AuthenticateWithAuth0(email string, password string) (string, error, string) {
	auth0URL := fmt.Sprintf("https://%s/oauth/token", r.auth0Domain)

	// Prepare the request payload
	requestBody, err := json.Marshal(map[string]interface{}{
		"grant_type":    "password",
		"username":      email,
		"password":      password,
		"client_id":     r.clientID,
		"client_secret": r.clientSecret,
		"audience":      r.audience,             // Ensure this matches your API identifier
		"scope":         "openid profile email", // Ensure proper scopes
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err), ""
	}

	resp, err := http.Post(auth0URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error sending request to Auth0: %w", err), ""
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err), ""
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Auth0 authentication failed: %s", string(body)), ""
	}

	var tokenResponse models.Auth0TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling token response: %w", err), ""
	}

	// Log the entire token response for debugging purposes
	log.Printf("Auth0 response for user %s: %+v", email, tokenResponse)

	// Extract Auth0 ID (sub claim) from ID Token
	auth0ID, userEmail, err := extractAuth0IDAndEmail(tokenResponse.IDToken)
	if err != nil {
		return "", fmt.Errorf("error extracting Auth0 ID: %w", err), ""
	}

	fmt.Println("✅ Auth0 ID:", auth0ID, "Email: ", userEmail)

	// Return the Auth0 ID
	return auth0ID, nil, userEmail
}

// RegisterWithAuth0 registers a new user with Auth0
func (r *AuthRepository) RegisterWithAuth0(email string, password string) (string, error) {
	auth0URL := fmt.Sprintf("https://%s/api/v2/users", r.auth0Domain)

	// Obtain a Management API token
	managementToken, err := r.getManagementAPIToken()
	if err != nil {
		return "", fmt.Errorf("failed to obtain management API token: %w", err)
	}

	// Prepare the request payload
	requestBody, err := json.Marshal(map[string]interface{}{
		"email":      email,
		"password":   password,
		"connection": "Username-Password-Authentication",
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest("POST", auth0URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", managementToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to Auth0: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Auth0 registration failed: %s", string(body))
	}

	var userResponse map[string]interface{}
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling user response: %w", err)
	}

	auth0ID, ok := userResponse["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("Auth0 ID not found in response")
	}

	return auth0ID, nil
}

// getManagementAPIToken obtains a Management API token from Auth0
func (r *AuthRepository) getManagementAPIToken() (string, error) {
	auth0URL := fmt.Sprintf("https://%s/oauth/token", r.auth0Domain)

	// Prepare the request payload
	requestBody, err := json.Marshal(map[string]interface{}{
		"grant_type":    "client_credentials",
		"client_id":     r.clientID,
		"client_secret": r.clientSecret,
		"audience":      fmt.Sprintf("https://%s/api/v2/", r.auth0Domain),
	})
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}

	resp, err := http.Post(auth0URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error sending request to Auth0: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Auth0 token request failed: %s", string(body))
	}

	var tokenResponse map[string]interface{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling token response: %w", err)
	}

	managementToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found in response")
	}

	return managementToken, nil
}

// extractAuth0IDAndEmail decodes the ID Token and extracts the `sub` (Auth0 ID) and `email` claims
func extractAuth0IDAndEmail(idToken string) (string, string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		auth0ID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)

		if auth0ID == "" {
			return "", "", fmt.Errorf("Auth0 ID not found in token")
		}
		if email == "" {
			return "", "", fmt.Errorf("Email not found in token")
		}

		return auth0ID, email, nil
	}

	return "", "", fmt.Errorf("Invalid token claims")
}
