package models

import "fmt"

type APIError struct {
	Code       string `json:"code"`              // Unique error code (e.g., "invalid_input", "not_found")
	Message    string `json:"message"`           // Human-readable error message
	Details    any    `json:"details,omitempty"` // Optional: Additional details (e.g., field validation errors)
	StatusCode int    `json:"-"`                 // HTTP status code (e.g., 400, 404, 500)
}

func (e APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewAPIError(code string, message string, details any) APIError {
	return APIError{Code: code, Message: message, Details: details}
}
