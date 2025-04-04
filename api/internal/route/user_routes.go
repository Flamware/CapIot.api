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

	// Get all users endpoint
	r.Handle("/api/users", middleware.JWTMiddleware(http.HandlerFunc(userHandler.GetAllUsers))).Methods(http.MethodGet)

	// Get current user
	r.Handle("/api/users/me", middleware.JWTMiddleware(http.HandlerFunc(userHandler.GetCurrentUser))).Methods(http.MethodGet)

	// Update current user
	r.Handle("/api/users/me", middleware.JWTMiddleware(http.HandlerFunc(userHandler.UpdateCurrentUser))).Methods(http.MethodPatch)
}
