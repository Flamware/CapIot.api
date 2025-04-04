// internal/handlers/user_handler.go
package handlers

import (
	"CapIot-api/internal/service"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
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

	var requestBody struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}

	user.Name = requestBody.Name
	updatedUser, err := h.userService.UpdateUser(*user)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedUser); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// GetAllUsers handles the request to retrieve all users.
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
