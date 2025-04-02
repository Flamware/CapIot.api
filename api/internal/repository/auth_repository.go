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
