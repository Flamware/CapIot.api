package route

import (
	"CapIot-api/internal/handlers"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupAuthRoutes initializes the authentication routes using the AuthHandler
func SetupAuthRoutes(r *mux.Router, authHandler *handlers.AuthHandler) {
	r.HandleFunc("/api/login", authHandler.LoginHandler).Methods(http.MethodPost)
}
