package server

import (
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// ValidationError represents a validation error with detailed information
type ValidationError struct {
	Message string
	Details []messagingapi.ErrorDetail
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}

// ToErrorResponse converts ValidationError to messagingapi.ErrorResponse
func (e *ValidationError) ToErrorResponse() messagingapi.ErrorResponse {
	response := messagingapi.ErrorResponse{
		Message: e.Message,
	}
	if len(e.Details) > 0 {
		response.Details = &e.Details
	}
	return response
}

// NewValidationError creates a new ValidationError with a message
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		Message: message,
		Details: []messagingapi.ErrorDetail{},
	}
}

// AddDetail adds a detail to the validation error
func (e *ValidationError) AddDetail(message, property string) {
	detail := messagingapi.ErrorDetail{}
	if message != "" {
		detail.Message = &message
	}
	if property != "" {
		detail.Property = &property
	}
	e.Details = append(e.Details, detail)
}