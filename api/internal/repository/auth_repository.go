package repository

import (
	"CapIot-api/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/auth0/go-auth0/management"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type AuthRepository struct {
	auth0Domain     string
	clientID        string
	clientSecret    string
	audience        string
	mgmt            *management.Management
	cachedToken     string
	tokenExpiryTime time.Time
	mu              sync.Mutex
	tokenFile       string
}

func NewAuthRepository() (*AuthRepository, error) {
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	audience := fmt.Sprintf("https://%s/api/v2/", auth0Domain)

	// Get the directory of the current source file (assuming this code is in repository/auth_repository.go)
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	dir := filepath.Dir(filename)

	// Define the token file path (same directory as the source file)
	tokenFilePath := filepath.Join(dir, "authToken.json")

	// Create the repository object
	repo := &AuthRepository{
		auth0Domain:  auth0Domain,
		clientID:     clientID,
		clientSecret: clientSecret,
		audience:     audience,
		tokenFile:    tokenFilePath,
	}

	// Attempt to load the token from the file
	err := repo.LoadTokenFromFile()
	if err != nil {
		log.Println("ℹ️ No valid token found in file, requesting a new one...")
		token, err := repo.getManagementAPIToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get management API token: %w", err)
		}
		repo.cachedToken = token
		repo.saveTokenToFile() // Save the newly fetched token
	} else {
		log.Println("✅ Token loaded from file.")
	}
	// Log token expiry time
	log.Printf("ℹ️ Token will expire at: %s", repo.tokenExpiryTime.Format(time.RFC3339))

	// Initialize the management client with the token
	mgmt, err := management.New(repo.auth0Domain, management.WithStaticToken(repo.cachedToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create auth0 management client: %w", err)
	}
	repo.mgmt = mgmt

	return repo, nil
}

// getManagementAPIToken returns a cached Auth0 Management token or fetches a new one
func (r *AuthRepository) getManagementAPIToken() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cachedToken != "" && time.Now().Before(r.tokenExpiryTime) {
		log.Println("Using cached Auth0 Management token")
		return r.cachedToken, nil
	}

	// Fetch a new token
	auth0URL := fmt.Sprintf("https://%s/oauth/token", r.auth0Domain)

	payload, err := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     r.clientID,
		"client_secret": r.clientSecret,
		"audience":      fmt.Sprintf("https://%s/api/v2/", r.auth0Domain),
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal token request: %w", err)
	}

	resp, err := http.Post(auth0URL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to request management token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to obtain management token: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"` // In seconds
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	// Cache the token and its expiry time
	r.cachedToken = tokenResp.AccessToken
	r.tokenExpiryTime = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // buffer 60s
	r.cachedToken = tokenResp.AccessToken
	r.tokenExpiryTime = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second) // buffer 60s
	log.Printf("ℹ️ Token will expire at: %s", r.tokenExpiryTime.Format(time.RFC3339))
	// Save token to file for persistence
	err = r.saveTokenToFile()
	if err != nil {
		log.Printf("Warning: Failed to save token to file: %v", err)
	}

	// Reinitialize the management client with the new token
	client, err := management.New(r.auth0Domain, management.WithStaticToken(r.cachedToken))
	if err != nil {
		return "", fmt.Errorf("failed to initialize Auth0 Management API client: %w", err)
	}
	r.mgmt = client

	return r.cachedToken, nil
}

// Save token to file
func (r *AuthRepository) saveTokenToFile() error {
	tokenData := struct {
		Token     string    `json:"access_token"`
		ExpiresAt time.Time `json:"expires_at"`
	}{
		Token:     r.cachedToken,
		ExpiresAt: r.tokenExpiryTime,
	}

	// Use os.WriteFile instead of ioutil for Go 1.16+
	data, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	err = os.WriteFile(r.tokenFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to save token to file: %w", err)
	}

	return nil
}

func (r *AuthRepository) LoadTokenFromFile() error {
	// Check if the token file exists
	data, err := os.ReadFile(r.tokenFile)
	if err != nil {
		log.Printf("ℹ️ Error reading token file '%s': %v", r.tokenFile, err)
		return err
	}

	var tokenData struct {
		Token     string    `json:"access_token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		log.Printf("⚠️ Error unmarshalling token data from '%s': %v", r.tokenFile, err)
		return err
	}

	// If the token is expired, return an error
	if time.Now().After(tokenData.ExpiresAt) {
		log.Printf("⚠️ Token in '%s' has expired.", r.tokenFile)
		r.cachedToken = "" // Invalidate cached token
		return fmt.Errorf("token has expired")
	}

	// Otherwise, set the cached token and expiration time
	r.cachedToken = tokenData.Token
	r.tokenExpiryTime = tokenData.ExpiresAt

	log.Println("✅ Token loaded from file")
	return nil
}

// AuthenticateWithAuth0 verifies user credentials via Auth0 and returns user info with roles
func (r *AuthRepository) AuthenticateWithAuth0(email, password string) (*models.AuthResult, error) {
	auth0URL := fmt.Sprintf("https://%s/oauth/token", r.auth0Domain)

	requestBody, err := json.Marshal(map[string]interface{}{
		"grant_type":    "password",
		"username":      email,
		"password":      password,
		"client_id":     r.clientID,
		"client_secret": r.clientSecret,
		"audience":      r.audience,
		"scope":         "openid profile email",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	resp, err := http.Post(auth0URL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed: %s", string(body))
	}

	var tokenResponse models.Auth0TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}

	auth0ID, userEmail, err := extractAuth0IDAndEmail(tokenResponse.IDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to extract Auth0 ID and email: %w", err)
	}

	roles, err := r.GetUserRolesByAuth0ID(auth0ID)
	if err != nil {
		log.Printf("⚠️ Could not retrieve roles for user %s: %v", auth0ID, err)
		roles = []string{}
	}

	return &models.AuthResult{Auth0ID: auth0ID, Email: userEmail, Role: roles}, nil
}

// RegisterWithAuth0 creates a new user in Auth0
func (r *AuthRepository) RegisterWithAuth0(email, password string) (string, error) {
	err := r.ensureMgmtClient()
	if err != nil {
		return "", err
	}
	conn := "Username-Password-Authentication"
	user := &management.User{
		Connection: &conn,
		Email:      &email,
		Password:   &password,
	}

	err = r.mgmt.User.Create(context.Background(), user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	log.Printf("✅ User registered with Auth0 ID: %s", *user.ID)
	return *user.ID, nil
}

// GetRoleIDByName returns the role ID from Auth0 by role name
func (r *AuthRepository) GetRoleIDByName(roleName string) (string, error) {
	err := r.ensureMgmtClient()
	if err != nil {
		return "", err
	}

	roles, err := r.mgmt.Role.List(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to list roles: %w", err)
	}

	for _, role := range roles.Roles {
		if role.Name != nil && *role.Name == roleName {
			return *role.ID, nil
		}
	}
	return "", fmt.Errorf("role '%s' not found", roleName)
}

// ensureMgmtClient guarantees that the management client is valid
func (r *AuthRepository) ensureMgmtClient() error {
	_, err := r.getManagementAPIToken()
	return err
}

// extractAuth0IDAndEmail decodes the ID token and extracts sub/email
func extractAuth0IDAndEmail(idToken string) (string, string, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", "", fmt.Errorf("token parse error: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		auth0ID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		if auth0ID == "" || email == "" {
			return "", "", fmt.Errorf("missing sub or email in token")
		}
		return auth0ID, email, nil
	}
	return "", "", fmt.Errorf("invalid token claims")
}

// GetUserRolesByAuth0ID retrieves the roles of a user by their Auth0 user ID
func (r *AuthRepository) GetUserRolesByAuth0ID(auth0UserID string) ([]string, error) {
	// Ensure the management client is initialized and the token is valid
	err := r.ensureMgmtClient()
	if err != nil {
		return nil, fmt.Errorf("auth0 management client is not initialized: %w", err)
	}

	// Retrieve roles associated with the user by their Auth0 ID
	roles, err := r.mgmt.User.Roles(context.Background(), auth0UserID)
	if err != nil {
		return nil, fmt.Errorf("GetUserRolesByAuth0ID: failed to retrieve roles for user %s: %w", auth0UserID, err)
	}

	// Collect role names
	var roleNames []string
	for _, role := range roles.Roles {
		if role.Name != nil {
			roleNames = append(roleNames, *role.Name)
		}
	}

	log.Printf("✅ User %s has the following roles: %v", auth0UserID, roleNames)
	return roleNames, nil
}

// SetUserRoleInAuth0 assigns a role to a user in Auth0
func (r *AuthRepository) SetUserRoleInAuth0(auth0UserID, roleID string) error {
	err := r.ensureMgmtClient()
	if err != nil {
		return err
	}

	err = r.mgmt.User.AssignRoles(context.Background(), auth0UserID, []*management.Role{
		{ID: &roleID},
	})
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	log.Printf("✅ Role %s assigned to user %s", roleID, auth0UserID)
	return nil
}
