package route

import (
	"api.cap.iot/service"
	"net/http"
)

// SetupRouter initializes the routes for the application
func SetupRouter(authService *service.AuthService) *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a public endpoint"))
	})

	// Setup user routes with the AuthRepository passed in
	SetupAuthRoutes(mux, authService)
	SetupLocationRoutes(mux)

	return mux
}
