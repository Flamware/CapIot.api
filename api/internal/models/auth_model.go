package models

// LoginRequest contains the data coming from the client for login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Auth0UserInfo contains the Auth0 user information
type Auth0UserInfo struct {
	Sub string `json:"sub"` // This is the Auth0 user ID.
	// Add other fields you need...
}

// Auth0TokenResponse contains the response from Auth0 token endpoint
type Auth0TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}
