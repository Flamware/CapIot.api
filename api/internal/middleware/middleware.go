package middleware

import (
	"CapIot-api/internal/auth" // Assuming your JWT validation is here
	"CapIot-api/internal/service"
	"context"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
)

// UserContextKey is a key for storing user information in the request context.
const UserClaimsContextKey = "user_claims"

// RoleContextKey is a key for storing user roles in the request context.
const RoleContextKey = "roles"

// JWTAuthMiddleware verifies the JWT and adds the user ID to the context.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("JWTAuthMiddleware: Starting JWT authentication")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("JWTAuthMiddleware: Authorization header missing")
			log.Printf("JWTAuthMiddleware: Request URL: %s, Method: %s", r.URL.Path, r.Method)
			log.Printf("JWTAuthMiddleware: Headers: %v", r.Header)
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Printf("JWTAuthMiddleware: Invalid token format, missing 'Bearer ': %s", authHeader)
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		log.Printf("JWTAuthMiddleware: Extracted token string: %s", tokenString)

		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("JWTAuthMiddleware: Invalid token: %s, error: %v", tokenString, err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Store the entire claims map in the request context
		ctx := context.WithValue(r.Context(), UserClaimsContextKey, claims)
		log.Printf("JWTAuthMiddleware: All claims stored in context with key '%s'", UserClaimsContextKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleCheckMiddleware(authService service.AuthService, requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("RoleCheckMiddleware: Checking role '%s'", requiredRole)

			// Retrieve the Auth0 User ID from the context
			auth0UserID := r.Context().Value(UserClaimsContextKey).(jwt.MapClaims)
			if auth0UserID == nil {
				log.Println("RoleCheckMiddleware: Auth0 User ID not found in context. Authentication likely failed.")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			auth0ID, ok := auth0UserID["sub"].(string)
			if !ok {
				log.Printf("RoleCheckMiddleware: Invalid Auth0 User ID type in context: %T, expected string", auth0UserID)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			log.Printf("RoleCheckMiddleware: Auth0 User ID found in context: %s", auth0ID)

			// Use the Auth0 User ID to retrieve roles
			roles, err := authService.GetUserRoles(r.Context(), auth0ID)
			if err != nil {
				log.Printf("RoleCheckMiddleware: Failed to retrieve roles for Auth0 User ID %s: %v", auth0ID, err)
				http.Error(w, "Failed to retrieve user roles", http.StatusInternalServerError)
				return
			}
			log.Printf("RoleCheckMiddleware: Retrieved roles for Auth0 User ID %s: %v", auth0ID, roles)

			hasRequiredRole := false
			for _, role := range roles {
				if role == requiredRole {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				log.Printf("RoleCheckMiddleware: Auth0 User ID %s does not have the required role '%s'.", auth0ID, requiredRole)
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			log.Printf("RoleCheckMiddleware: Auth0 User ID %s has the required role '%s'. Proceeding.", auth0ID, requiredRole)

			// Optionally, you can add the roles to the context for later use in handlers
			ctx := context.WithValue(r.Context(), RoleContextKey, roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
