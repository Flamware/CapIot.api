package main

import (
	"api.cap.iot/config"
	"api.cap.iot/route"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type TokenInfo struct {
	Token  string `json:"token"`
	Expiry string `json:"expiry"`
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// Connect to MongoDB
	log.Println("🔌 Initializing MongoDB connection...")
	config.ConnectDB()

	// Auth0 Configuration
	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	auth0Audience := os.Getenv("AUTH0_AUDIENCE")
	clientID := os.Getenv("AUTH0_CLIENT_ID")
	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	tokenInfoJSON := os.Getenv("TOKEN_INFO")

	var tokenInfo TokenInfo
	if tokenInfoJSON != "" {
		err = json.Unmarshal([]byte(tokenInfoJSON), &tokenInfo)
		if err != nil {
			log.Fatalf("❌ Failed to parse token info: %v", err)
		}
	}

	// Check if the token is expired
	if tokenInfo.Token == "" || isTokenExpired(tokenInfo.Expiry) {
		urlStr := "https://" + auth0Domain + "/oauth/token"

		payload := strings.NewReader(fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s&audience=https://%s/api/v2/",
			clientID, clientSecret, auth0Domain))

		req, err := http.NewRequest("POST", urlStr, payload)
		if err != nil {
			log.Fatalf("❌ Failed to create request: %v", err)
		}

		req.Header.Add("content-type", "application/x-www-form-urlencoded")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("❌ Failed to send request: %v", err)
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("❌ Failed to read response body: %v", err)
		}

		log.Println(res)
		log.Println(string(body))

		// Unmarshal the response body
		var tokenResponse map[string]interface{}
		err = json.Unmarshal(body, &tokenResponse)
		if err != nil {
			log.Fatalf("❌ Failed to parse token response: %v", err)
		}

		// Check for access_token in the response
		managementToken, ok := tokenResponse["access_token"].(string)
		if !ok || managementToken == "" {
			log.Fatal("❌ Failed to retrieve management token")
		} else {
			log.Println("✅ Management token retrieved successfully")
			// Calculate token expiry time (assuming token is valid for 24 hours)
			expiryTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
			tokenInfo = TokenInfo{
				Token:  managementToken,
				Expiry: expiryTime,
			}
			// Marshal the token info to JSON and store it in the .env file
			tokenInfoBytes, err := json.Marshal(tokenInfo)
			if err != nil {
				log.Fatalf("❌ Failed to marshal token info: %v", err)
			}
			f, err := os.OpenFile(".env", os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				log.Fatalf("❌ Failed to open .env file: %v", err)
			}
			defer f.Close()

			if _, err = f.WriteString(fmt.Sprintf("\nTOKEN_INFO=%s", string(tokenInfoBytes))); err != nil {
				log.Fatalf("❌ Failed to write token info to .env file: %v", err)
			}
		}
	} else {
		log.Println("✅ Using existing management token")
	}

	// Ensure necessary environment variables are set
	if auth0Domain == "" || auth0Audience == "" {
		log.Fatal("❌ AUTH0_DOMAIN, AUTH0_AUDIENCE, or AUTH0_MANAGEMENT_TOKEN is missing")
	}

	mux := route.SetupRouter()

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	// Start the server
	log.Println("🚀 Starting server on :8080...")
	err = http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func isTokenExpired(expiry string) bool {
	expiryTime, err := time.Parse(time.RFC3339, expiry)
	if err != nil {
		log.Fatalf("❌ Failed to parse token expiry time: %v", err)
	}
	return time.Now().After(expiryTime)
}
