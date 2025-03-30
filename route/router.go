package route

import (
	"api.cap.iot/middleware"
	"net/http"
)

// SetupRouter initializes the routes for the application
func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoint
	mux.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a public endpoint"))
	})

	// Admin-only endpoint
	mux.Handle("/admin", middleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome, admin!"))
	})))

	SetupUserRoutes(mux)
	SetupLocationRoutes(mux)

	return mux
}
