package utils

import (
	"CapIot-api/internal/models"
	"encoding/json"
	"log"
	"net/http"
)

// RespondWithError sends a JSON error response.
func RespondWithError(writer http.ResponseWriter, statusCode int, message string, details interface{}) {
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	errResponse := models.APIError{
		Code:    GetStatusText(statusCode), // Use the general GetStatusText
		Message: message,
		Details: details,
	}
	if err := json.NewEncoder(writer).Encode(errResponse); err != nil {
		log.Printf("Failed to encode error response: %v", err)
		http.Error(writer, "Failed to send error response", http.StatusInternalServerError)
	}
}

// GetStatusText returns a generic error code based on the HTTP status code.
func GetStatusText(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusInternalServerError:
		return "internal_server_error"
	default:
		return "error"
	}
}

// RespondWithJSON sends a JSON success response. You might want a similar function for success.
func RespondWithJSON(writer http.ResponseWriter, statusCode int, payload interface{}) {
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		http.Error(writer, "Failed to send JSON response", http.StatusInternalServerError)
	}
}
