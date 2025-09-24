package middleware

import (
	"CapIot-api/internal/auth"
	"CapIot-api/internal/config"
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/models"
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

// JWTAuthMiddleware verifies the JWT and adds the user ID and roles to the context.
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

		ctx := context.WithValue(r.Context(), config.UserClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// hasAdminRole checks if the user has the "admin" role.
func hasAdminRole(claims jwt.MapClaims) bool {
	if roles, ok := claims["role"].([]interface{}); ok {
		for _, role := range roles {
			if r, ok := role.(string); ok && r == "admin" {
				return true
			}
		}
	}
	return false
}

// RoleCheckMiddleware checks if the authenticated user has at least one of the required roles.
func RoleCheckMiddleware(authService service.AuthService, requiredRoles []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("RoleCheckMiddleware: Checking required roles %v", requiredRoles)

			claims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				log.Println("RoleCheckMiddleware: Claims not found or invalid in context.")
				apiErr := models.NewAPIError(models.ErrorCodeUnauthorized, "Unauthorized", nil, http.StatusUnauthorized)
				utils.RespondWithError(w, apiErr)
				return
			}
			// Allow if the user is an admin
			if hasAdminRole(claims) {
				log.Println("RoleCheckMiddleware: User is an admin. Proceeding.")
				next.ServeHTTP(w, r)
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

			userRoles, err := authService.GetUserRoles(r.Context(), auth0ID)
			if err != nil {
				log.Printf("RoleCheckMiddleware: Failed to retrieve roles for Auth0 User ID %s: %v", auth0ID, err)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve user roles", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("RoleCheckMiddleware: Retrieved roles for Auth0 User ID %s: %v", auth0ID, userRoles)

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

			ctx := context.WithValue(r.Context(), config.RoleContextKey, userRoles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CheckLocationAccess checks if the user has access to a specific site based on a site ID.
func CheckLocationAccess(locationHandler *handlers.LocationHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			// Allow if the user is an admin
			if hasAdminRole(userClaims) {
				log.Println("CheckLocationAccess: User is an admin. Proceeding.")
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := userClaims["id"].(float64)
			userIDInt := int(userID)

			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			vars := mux.Vars(r)
			locationIDStr := vars["locationID"]
			if locationIDStr == "" {
				log.Println("CheckLocationAccess: 'locationID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"locationID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}
			locationID, err := strconv.ParseInt(locationIDStr, 10, 64)
			if err != nil {
				log.Printf("CheckLocationAccess: Invalid locationID format: %v", err)
				apiErr := models.NewAPIError(models.ErrorCodeInvalidFormat, "Invalid locationID format", nil, http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			hasAccess, err := locationHandler.CheckLocationAccess(userIDInt, locationID)
			if err != nil {
				log.Printf("CheckLocationAccess: Error checking access for User ID %d to location %d: %v", userIDInt, locationID, err)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error checking access permissions", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			if !hasAccess {
				log.Printf("CheckLocationAccess: User ID %d does not have access to location %d", userIDInt, locationID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this location", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CheckDeviceAccess checks if the user has access to a specific device.
func CheckDeviceAccess(deviceHandler *handlers.DeviceHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			// Allow if the user is an admin
			if hasAdminRole(userClaims) {
				log.Println("CheckDeviceAccess: User is an admin. Proceeding.")
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := userClaims["id"].(float64)
			userIDInt := int(userID)

			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			vars := mux.Vars(r)
			deviceID := vars["deviceID"]
			if deviceID == "" {
				log.Println("CheckDeviceAccess: 'deviceID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"deviceID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			hasAccess := deviceHandler.CheckDeviceAccess(userIDInt, deviceID)
			if !hasAccess {
				log.Printf("CheckDeviceAccess: User ID %d does not have access to device %s", userIDInt, deviceID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this device", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("CheckDeviceAccess: User %d has access to device %s. Proceeding.", userIDInt, deviceID)
			next.ServeHTTP(w, r)
		})
	}
}

// CheckDeviceRights checks if a device token has access to a specific device.
func CheckDeviceRights(handler *handlers.DeviceHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This middleware is for device-to-device/gateway communication, not user access.
			// It doesn't need to check for the admin role from user claims.
			vars := mux.Vars(r)
			deviceID := vars["deviceID"]
			if deviceID == "" {
				log.Println("CheckDeviceRights: 'deviceID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"deviceID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			tokenHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenHeader, "Bearer ")
			if token == tokenHeader {
				log.Printf("CheckDeviceRights: Invalid token format, missing 'Bearer ': %s", tokenHeader)
				apiErr := models.NewAPIError(models.ErrorCodeInvalidToken, "Invalid token format, missing 'Bearer '", nil, http.StatusUnauthorized)
				utils.RespondWithError(w, apiErr)
				return
			}

			allowed := handler.CheckDeviceRights(token, deviceID)
			if !allowed {
				apiError := models.NewAPIError(models.ErrorCodeForbidden, "Access denied for device", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiError)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CheckDeviceLocationRights checks if a device token has access to a device and if it belongs to a location.
func CheckDeviceLocationRights(handler *handlers.DeviceHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			deviceID := vars["deviceID"]
			if deviceID == "" {
				log.Println("CheckDeviceLocationRights: 'deviceID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"deviceID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return // Add return here
			}
			locationID := vars["locationID"]
			if locationID == "" {
				log.Println("CheckDeviceLocationRights: 'locationID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"locationID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return // Add return here
			}

			tokenHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(tokenHeader, "Bearer ")
			if token == tokenHeader {
				log.Printf("CheckDeviceLocationRights: Invalid token format, missing 'Bearer ': %s", tokenHeader)
				apiErr := models.NewAPIError(models.ErrorCodeInvalidToken, "Invalid token format, missing 'Bearer '", nil, http.StatusUnauthorized)
				utils.RespondWithError(w, apiErr)
				return // Add return here
			}

			allowed := handler.CheckDeviceRights(token, deviceID)
			if !allowed {
				apiError := models.NewAPIError(models.ErrorCodeForbidden, "Access denied for device", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiError)
				return // Add return here
			}

			allowed = handler.CheckDeviceLocation(deviceID, locationID)
			if !allowed {
				apiError := models.NewAPIError(models.ErrorCodeForbidden, "Device does not belong to the specified location", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiError)
				return // Add return here
			}

			log.Printf("CheckDeviceLocationRights: User has rights for device %s and location %s. Proceeding.", deviceID, locationID)
			next.ServeHTTP(w, r)
		})
	}
}

// CheckSiteAccess checks if the user has access to a specific site.
func CheckSiteAccess(location *handlers.LocationHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			// Allow if the user is an admin
			if hasAdminRole(userClaims) {
				log.Println("CheckSiteAccess: User is an admin. Proceeding.")
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := userClaims["id"].(float64)
			userIDInt := int(userID)

			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			vars := mux.Vars(r)
			siteIDStr := vars["siteID"]
			if siteIDStr == "" {
				log.Println("CheckSiteAccess: 'siteID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"siteID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}
			siteID, err := strconv.ParseInt(siteIDStr, 10, 64)
			if err != nil {
				log.Printf("CheckSiteAccess: Invalid siteID format: %v", err)
				apiErr := models.NewAPIError(models.ErrorCodeInvalidFormat, "Invalid siteID format", nil, http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			hasAccess, err := location.CheckSiteAccess(userIDInt, siteID)
			if err != nil {
				log.Printf("CheckSiteAccess: Error checking access for User ID %d to site %d: %v", userIDInt, siteID, err)
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Error checking access permissions", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			if !hasAccess {
				log.Printf("CheckSiteAccess: User ID %d does not have access to site %d", userIDInt, siteID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this site", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}
			log.Printf("CheckSiteAccess: User %d has access to site %d. Proceeding.", userIDInt, siteID)
			next.ServeHTTP(w, r)
		})
	}
}

// CheckDeviceLocationRights checks if a device token has access to a device and if it belongs to a location.
func CheckUserRights(handler *handlers.DeviceHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
			if !ok {
				log.Printf("CheckUserRights: Claims not found or invalid in context.")
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}
			// Allow if the user is an admin
			if hasAdminRole(userClaims) {
				log.Println("CheckLocationAccess: User is an admin. Proceeding.")
				next.ServeHTTP(w, r)
				return
			}

			userID, ok := userClaims["id"].(float64)
			userIDInt := int(userID)
			if !ok {
				apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
				utils.RespondWithError(w, apiErr)
				return
			}

			vars := mux.Vars(r)
			deviceID := vars["deviceID"]
			if deviceID == "" {
				log.Println("CheckUserRights: 'deviceID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"deviceID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}
			locationID := vars["locationID"]
			if locationID == "" {
				log.Println("CheckUserRights: 'locationID' parameter missing in URL.")
				apiErr := models.NewAPIError(
					models.ErrorCodeMissingParameter,
					"locationID parameter missing",
					nil,
					http.StatusBadRequest)
				utils.RespondWithError(w, apiErr)
				return
			}

			hasAccess := handler.CheckDeviceAccess(userIDInt, deviceID)
			if !hasAccess {
				log.Printf("CheckUserRights: User ID %d does not have access to device %s", userIDInt, deviceID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Insufficient permissions for this device", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}

			hasAccess = handler.CheckDeviceLocation(deviceID, locationID)
			if !hasAccess {
				log.Printf("CheckUserRights: Device %s does not belong to location %s", deviceID, locationID)
				apiErr := models.NewAPIError(models.ErrorCodeInsufficientPermissions, "Device does not belong to the specified location", nil, http.StatusForbidden)
				utils.RespondWithError(w, apiErr)
				return
			}

			log.Printf("CheckUserRights: User %d has rights for device %s and location %s. Proceeding.", userIDInt, deviceID, locationID)
			next.ServeHTTP(w, r)
		})
	}
}
