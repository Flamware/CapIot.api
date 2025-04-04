package route

import (
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/service"
	"encoding/json"
	"net/http"
)

func SetupUserRoutes(mux *http.ServeMux, userService *service.DefaultUserService) {
	// Example location endpoint
	mux.HandleFunc("/api/bind-device/{deviceID}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is a location endpoint"))
	})

	// Get users endpoint
	mux.Handle("/api/users", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		users, err := userService.GetAllUsers()
		if err != nil {
			http.Error(w, "Failed to get users", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})))
}
