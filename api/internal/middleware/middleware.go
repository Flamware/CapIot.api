package middleware

import (
	"CapIot-api/internal/auth"
	"CapIot-api/internal/config"
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/models" // Import the models package
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// JWTAuthMiddleware verifies the JWT and adds the user ID to the context.
func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("JWTAuthMiddleware: Authorization header missing")
			apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Authorization header missing", nil, http.StatusUnauthorized)
			utils.RespondWithError(w, apiErr)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			log.Printf("JWTAuthMiddleware: Invalid token format, missing 'Bearer ': %s", authHeader)
			apiErr := models.NewAPIError(models.ErrorCodeInvalidToken, "Invalid token format, missing 'Bearer '", nil, http.StatusUnauthorized)
			utils.RespondWithError(w, apiErr)
			return
		}

		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			log.Printf("JWTAuthMiddleware: Invalid token: %s, error: %v", tokenString, err)
			apiErr := models.NewAPIError(models.ErrorCodeInvalidToken, "Invalid token", nil, http.StatusUnauthorized)
			utils.RespondWithError(w, apiErr)
			return
		}

		// Store the entire claims map in the request context
		ctx := context.WithValue(r.Context(), config.UserClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleCheckMiddleware checks if the authenticated user has at least one of the required roles.
func RoleCheckMiddleware(authService service.AuthService, requiredRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("RoleCheckMiddleware: Checking required roles %v", requiredRoles)

			// Retrieve the Auth0 User ID from the context
			auth0UserID := r.Context().Value(config.UserClaimsContextKey)
			if auth0UserID == nil {
				log.Println("RoleCheckMiddleware: Auth0 User ID not found in context. Authentication likely failed.")
				apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
				utils.RespondWithError(w, apiErr)
				return
			}

			claims, ok := auth0UserID.(jwt.MapClaims)
			if !ok {
				log.Printf("RoleCheckMiddleware: Invalid Auth0 claims type in context: %T, expected jwt.MapClaims", auth0UserID)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal Server Error", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			auth0ID, ok := claims["sub"].(string)
			if !ok {
				log.Printf("RoleCheckMiddleware: Invalid Auth0 'sub' type in claims: %T, expected string", claims["sub"])
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal Server Error", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("RoleCheckMiddleware: Auth0 User ID found in context: %s", auth0ID)

			// Use the Auth0 User ID to retrieve roles
			userRoles, err := authService.GetUserRoles(r.Context(), auth0ID)
			if err != nil {
				log.Printf("RoleCheckMiddleware: Failed to retrieve roles for Auth0 User ID %s: %v", auth0ID, err)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve user roles", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("RoleCheckMiddleware: Retrieved roles for Auth0 User ID %s: %v", auth0ID, userRoles)

			// Check if the user's roles contain at least one of the required roles
			hasRequiredRole := false
			for _, requiredRole := range requiredRoles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRequiredRole = true
						break
					}
				}
				if hasRequiredRole {
					break
				}
			}

			if !hasRequiredRole {
				log.Printf("RoleCheckMiddleware: Auth0 User ID %s does not have one of the required roles %v.", auth0ID, requiredRoles)
				apiErr := models.NewAPIError(models.ErrorCodeForbidden, "Insufficient permissions", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}

			log.Printf("RoleCheckMiddleware: Auth0 User ID %s has one of the required roles. Proceeding.", auth0ID)

			// Optionally, you can add the roles to the context for later use in handlers
			ctx := context.WithValue(r.Context(), config.RoleContextKey, userRoles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CheckSiteAccess checks if the user has access to a specific site based on a location ID.
func CheckSiteAccess(locationService service.LocationService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve claims from the context
			claims, ok := r.Context().Value(config.UserClaimsContextKey).(*utils.Claims)
			if !ok {
				log.Println("CheckSiteAccess: Claims not found or invalid type in context.")
				apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
				utils.RespondWithError(w, apiErr)
				return
			}

			userId := claims.ID
			if !ok {
				log.Println("CheckSiteAccess: 'sub' (Auth0 User ID) not found or is not a string.")
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Internal Server Error", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			// Extract locationID from URL parameters, its a int64
			vars := mux.Vars(r)
			locationIDStr := vars["locationId"]
			if locationIDStr == "" {
				log.Println("CheckSiteAccess: 'locationId' parameter missing in URL.")
				apiErr := models.NewAPIError(models.ErrorCodeMissingParameter, "locationId parameter missing", nil, http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}
			locationID, err := strconv.ParseInt(locationIDStr, 10, 64)
			if err != nil {
				log.Printf("CheckSiteAccess: Invalid locationId format: %v", err)
				apiErr := models.NewAPIError(models.ErrorCodeInvalidFormat, "Invalid locationId format", nil, http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			// Check if the user has access to the location's site
			hasAccess, err := locationService.CheckUserAccessToLocation(userId, locationID)
			if err != nil {
				log.Printf("CheckSiteAccess: Error checking access for Auth0 User ID %s to location %s: %v", userId, locationID, err)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error checking access permissions", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			if !hasAccess {
				log.Printf("CheckSiteAccess: Auth0 User ID %s does not have access to location %s", userId, locationID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this location", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}

			log.Printf("CheckSiteAccess: User %s has access to location %s. Proceeding.", userId, locationID)
			next.ServeHTTP(w, r)
		})
	}
}

// CheckLocationAccess checks if the user has access to a specific site based on a site ID.
func CheckLocationAccess(locationHandler *handlers.LocationHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve claims from the context
			// Extract user ID from JWT claims
			userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			userID, ok := userClaims["id"].(float64)

			// convert float64 to int
			userIDInt := int(userID)

			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			vars := mux.Vars(r)
			locationIDStr := vars["locationId"]
			if locationIDStr == "" {
				log.Println("CheckLocationAccess: 'locationId' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"locationId parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}
			locationID, err := strconv.ParseInt(locationIDStr, 10, 64)
			if err != nil {
				log.Printf("CheckLocationAccess: Invalid locationId format: %v", err)
				apiErr := models.NewAPIError(models.ErrorCodeInvalidFormat, "Invalid locationId format", nil, http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			// Check if the user has access to the location's site
			hasAccess, err := locationHandler.CheckLocationAccess(userIDInt, locationID)
			if err != nil {
				log.Printf("CheckLocationAccess: Error checking access for Auth0 User ID %s to location %s: %v", userIDInt, locationID, err)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error checking access permissions", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			if !hasAccess {
				log.Printf("CheckLocationAccess: Auth0 User ID %s does not have access to location %s", userIDInt, locationID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this location", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
func CheckDeviceAccess(deviceHandler *handlers.DeviceHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Retrieve claims from the context
			// Extract user ID from JWT claims
			userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			userID, ok := userClaims["id"].(float64)

			// convert float64 to int
			userIDInt := int(userID)

			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			vars := mux.Vars(r)
			deviceID := vars["deviceID"]
			if deviceID == "" {
				log.Println("checkDeviceAccess: 'deviceId' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"deviceId parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			// Check if the user has access to the location's site
			hasAccess := deviceHandler.CheckDeviceAccess(userIDInt, deviceID)
			if !hasAccess {
				log.Printf("checkDeviceAccess: Auth0 User ID %s does not have access to device %s", userIDInt, deviceID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this device", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("checkDeviceAccess: User %s has access to device %s. Proceeding.", userIDInt, deviceID)
			next.ServeHTTP(w, r)
		})
	}
}
