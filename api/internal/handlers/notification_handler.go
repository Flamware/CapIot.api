package handlers

import (
	"CapIot-api/internal/config"
	"CapIot-api/internal/models"
	"CapIot-api/internal/service"
	"CapIot-api/internal/utils"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	page, limit := 1, 10
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	notifications, err := h.notificationService.GetNotifications(r.Context(), userID, page, limit)
	if err != nil {
		log.Printf("Failed to get notifications: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve notifications", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		log.Printf("Failed to encode notifications: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to encode response", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
}

func (h *NotificationHandler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	vars := mux.Vars(r)
	notificationIDStr := vars["notificationID"]
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil || notificationID <= 0 {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid notification ID", nil, http.StatusBadRequest)
		utils.RespondWithError(w, apiErr)
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), notificationID, userID); err != nil {
		log.Printf("Failed to mark notification %d as read: %v", notificationID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to mark notification as read", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	if err := h.notificationService.MarkAllAsRead(r.Context(), userID); err != nil {
		log.Printf("Failed to mark all notifications as read for user %d: %v", userID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to mark all notifications as read", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) DeleteNotification(writer http.ResponseWriter, request *http.Request) {
	userClaims, ok := request.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
	}
	userID := int(userIDFloat)

	vars := mux.Vars(request)
	notificationIDStr := vars["notificationID"]
	notificationID, err := strconv.Atoi(notificationIDStr)
	if err != nil || notificationID <= 0 {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Invalid notification ID", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}
	if err := h.notificationService.DeleteNotification(request.Context(), notificationID, userID); err != nil {
		log.Printf("Failed to delete notification %d: %v", notificationID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to delete notification", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	userClaims, ok := r.Context().Value(config.UserClaimsContextKey).(jwt.MapClaims)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user claims", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userIDFloat, ok := userClaims["id"].(float64)
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Invalid user ID", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}
	userID := int(userIDFloat)

	if err := h.notificationService.DeleteAllNotifications(r.Context(), userID); err != nil {
		log.Printf("Failed to delete all notifications for user %d: %v", userID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to delete all notifications", nil, http.StatusInternalServerError)
		utils.RespondWithError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *NotificationHandler) GetDeviceNotification(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	deviceID, ok := vars["deviceID"]
	if !ok {
		apiErr := models.NewAPIError(models.ErrorCodeBadRequest, "Missing device ID in path", nil, http.StatusBadRequest)
		utils.RespondWithError(writer, apiErr)
		return
	}

	// Pagination parameters
	page, limit := 1, 10
	if p, err := strconv.Atoi(request.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(request.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	notifications, err := h.notificationService.GetDeviceNotifications(request.Context(), deviceID, page, limit)
	if err != nil {
		log.Printf("Failed to get notifications for device %d: %v", deviceID, err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to retrieve device notifications", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(notifications); err != nil {
		log.Printf("Failed to encode device notifications: %v", err)
		apiErr := models.NewAPIError(models.ErrorCodeInternalServerError, "Failed to encode response", nil, http.StatusInternalServerError)
		utils.RespondWithError(writer, apiErr)
		return
	}
}
