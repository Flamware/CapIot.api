package middleware

import (
	"api.cap.iot/repository"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

// Auth0UserIDKey is the context key for the Auth0 user ID
type Auth0UserIDKey struct{}

// EnsureValidToken is a middleware function to validate JWT tokens.
func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse(os.Getenv("AUTH0_DOMAIN"))
	if err != nil {
		log.Fatalf("Failed to parse the issuer URL: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the JWT validator: %v", err)
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("JWT validation error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid token."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract the validated claims and store the user ID in context
				claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
				if !ok {
					log.Println("Failed to extract validated claims")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{"message":"Invalid token."}`))
					return
				}
				ctx := context.WithValue(r.Context(), Auth0UserIDKey{}, claims.RegisteredClaims.Subject)
				next.ServeHTTP(w, r.WithContext(ctx))
			})).ServeHTTP(w, r)
		})
	}
}

// RequireRole ensures the user has the specified role
func RequireRole(requiredRole, managementToken, auth0Domain string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Auth0 user ID from context
			auth0ID, ok := r.Context().Value(Auth0UserIDKey{}).(string)
			if !ok {
				log.Println("Auth0 user ID not found in context")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message":"User ID not found."}`))
				return
			}

			// Fetch user from database
			user, err := repository.GetUserByAuth0ID(auth0ID)
			if err != nil {
				log.Printf("User not found for Auth0 ID: %s", auth0ID)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message":"User not found."}`))
				return
			}

			// Check if the required role exists
			for _, role := range user.Roles {
				if strings.EqualFold(role, requiredRole) {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"Insufficient role permissions."}`))
		})
	}
}
