package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupUserRoutes(r *mux.Router, userHandler *handlers.UserHandler) {
	// Example location endpoint (Consider moving this if it's not user-specific)
	r.HandleFunc("/api/bind-device/{deviceID}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a location endpoint"))
	})

	// Get current user
	r.Handle("/api/users/me", middleware.JWTAuthMiddleware(http.HandlerFunc(userHandler.GetCurrentUser))).Methods(http.MethodGet)

	// Get locations of a user
	r.Handle("/api/users/me/locations", middleware.JWTAuthMiddleware(http.HandlerFunc(userHandler.GetUserLocations))).Methods(http.MethodGet)

	// Get sites of a user
	r.Handle("/api/sites/me", middleware.JWTAuthMiddleware(http.HandlerFunc(userHandler.GetMySites))).Methods(http.MethodGet)

	// Update current user
	r.Handle("/api/users/me", middleware.JWTAuthMiddleware(http.HandlerFunc(userHandler.UpdateCurrentUser))).Methods(http.MethodPatch)

	// Assign user to location
	r.Handle("/api/assign-user/{userID}", middleware.JWTAuthMiddleware(http.HandlerFunc(userHandler.AsignUser))).Methods(http.MethodPost)
}
