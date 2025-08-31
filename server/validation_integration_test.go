package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

func TestValidationEndpointHTTPResponse(t *testing.T) {
	// Create a test server with custom error handler
	s := &server{}
	
	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		var validationErr *ValidationError
		if errors, ok := err.(*ValidationError); ok {
			validationErr = errors
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(validationErr.ToErrorResponse())
			return
		}
		// Default error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	
	handler := messagingapi.NewStrictHandlerWithOptions(s, nil, messagingapi.StrictHTTPServerOptions{
		ResponseErrorHandlerFunc: errorHandler,
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
	})

	t.Run("returns 400 with JSON error for validation failure", func(t *testing.T) {
		// Create request with invalid data (empty messages array)
		requestBody := messagingapi.ValidateMessageRequest{
			Messages: []messagingapi.Message{},
		}
		body, _ := json.Marshal(requestBody)
		
		req := httptest.NewRequest(http.MethodPost, "/v2/bot/message/validate/broadcast", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		
		// Use the handler directly
		handler.ValidateBroadcast(w, req)
		
		// Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		
		// Parse response body
		var errorResponse messagingapi.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		
		assert.Equal(t, "The request body has 1 error(s)", errorResponse.Message)
		assert.NotNil(t, errorResponse.Details)
		assert.Len(t, *errorResponse.Details, 1)
		assert.Equal(t, "Size must be between 1 and 5", *(*errorResponse.Details)[0].Message)
		assert.Equal(t, "messages", *(*errorResponse.Details)[0].Property)
	})

	t.Run("returns 200 for valid messages", func(t *testing.T) {
		// Create request with valid data
		requestBody := messagingapi.ValidateMessageRequest{
			Messages: []messagingapi.Message{
				{Type: "text"},
			},
		}
		body, _ := json.Marshal(requestBody)
		
		req := httptest.NewRequest(http.MethodPost, "/v2/bot/message/validate/broadcast", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		
		// Use the handler directly
		handler.ValidateBroadcast(w, req)
		
		// Check response
		assert.Equal(t, http.StatusOK, w.Code)
	})
}