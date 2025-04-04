package route

import (
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/service"
	"github.com/gorilla/mux"
	"net/http"
)

// SetupAdminRoutes sets up the admin-protected route
func SetupAdminRoutes(r *mux.Router, authService service.AuthService) {
	adminRouter := r.PathPrefix("/api/admin").Subrouter()

	// Apply JWTAuthMiddleware
	adminRouter.Use(middleware.JWTAuthMiddleware)

	// Apply RoleCheckMiddleware for "admin" role
	adminRouter.Use(middleware.RoleCheckMiddleware(authService, "admin"))

	// Define the handler for the /test route within the admin subrouter
	adminRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Admin access granted"))
	}).Methods(http.MethodGet)

}
