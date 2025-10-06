package route

import (
	"CapIot-api/internal/handlers"
	"CapIot-api/internal/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupNotificationsRoute(r *mux.Router, notificationHandler *handlers.NotificationHandler, deviceHandler *handlers.DeviceHandler) {
	// 1️⃣ Get all notifications
	r.Handle("/api/notifications", middleware.JWTAuthMiddleware(http.HandlerFunc(notificationHandler.GetNotifications))).Methods(http.MethodGet)

	// 2️⃣ Mark all notifications as read (specific route)
	r.Handle("/api/notifications/mark-all-read", middleware.JWTAuthMiddleware(http.HandlerFunc(notificationHandler.MarkAllAsRead))).Methods(http.MethodPatch)

	// 3️⃣ Delete all notifications (general route for /api/notifications with DELETE)
	r.Handle("/api/notifications/delete-all", middleware.JWTAuthMiddleware(http.HandlerFunc(notificationHandler.DeleteAllNotifications))).Methods(http.MethodDelete)

	// 4️⃣ Mark single notification as read (specific route with ID)
	r.Handle("/api/notifications/{notificationID}", middleware.JWTAuthMiddleware(http.HandlerFunc(notificationHandler.MarkNotificationAsRead))).Methods(http.MethodPatch)

	// 5️⃣ Delete a single notification
	r.Handle("/api/notifications/{notificationID}", middleware.JWTAuthMiddleware(http.HandlerFunc(notificationHandler.DeleteNotification))).Methods(http.MethodDelete)

	r.Handle("/api/notifications/device/{deviceID}", middleware.JWTAuthMiddleware(
		middleware.CheckDeviceAccess(deviceHandler)(
			http.HandlerFunc(notificationHandler.GetDeviceNotification)))).Methods(http.MethodGet)
}
