package route

import (
	"CapIot-api/internal/middleware"
	"CapIot-api/internal/service"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
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

	mux.Handle("/api/users/me", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := r.Context().Value("user").(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid user claims", http.StatusInternalServerError)
			return
		}
		userIDFloat, ok := userClaims["id"].(float64)
		if !ok {
			http.Error(w, "Invalid user ID", http.StatusInternalServerError)
			return
		}
		userID := int(userIDFloat)

		switch r.Method {
		case http.MethodGet:
			user, err := userService.GetUserByID(userID)
			if err != nil {
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)

		case http.MethodPatch:
			var requestBody struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			user, err := userService.GetUserByID(userID)
			if err != nil {
				http.Error(w, "Failed to get user", http.StatusInternalServerError)
				return
			}

			user.Name = requestBody.Name
			updatedUser, err := userService.UpdateUser(*user)
			if err != nil {
				http.Error(w, "Failed to update user", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
				http.Error(w, "Error encoding response", http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

}
