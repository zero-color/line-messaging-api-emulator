package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

func TestValidateBroadcast(t *testing.T) {
	s := &server{}

	t.Run("validates valid broadcast message", func(t *testing.T) {
		request := messagingapi.ValidateBroadcastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{Type: "text"},
				},
			},
		}

		response, err := s.ValidateBroadcast(context.Background(), request)
		require.NoError(t, err)
		assert.IsType(t, messagingapi.ValidateBroadcast200Response{}, response)
	})

	t.Run("returns error for empty messages", func(t *testing.T) {
		request := messagingapi.ValidateBroadcastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{},
			},
		}

		_, err := s.ValidateBroadcast(context.Background(), request)
		require.Error(t, err)
		
		var validationErr *ValidationError
		require.True(t, errors.As(err, &validationErr))
		assert.Equal(t, "The request body has 1 error(s)", validationErr.Message)
		assert.Len(t, validationErr.Details, 1)
		assert.Equal(t, "Size must be between 1 and 5", *validationErr.Details[0].Message)
		assert.Equal(t, "messages", *validationErr.Details[0].Property)
	})

	t.Run("returns error for too many messages", func(t *testing.T) {
		messages := make([]messagingapi.Message, 6)
		for i := range messages {
			messages[i] = messagingapi.Message{Type: "text"}
		}

		request := messagingapi.ValidateBroadcastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: messages,
			},
		}

		_, err := s.ValidateBroadcast(context.Background(), request)
		require.Error(t, err)
		
		var validationErr *ValidationError
		require.True(t, errors.As(err, &validationErr))
		assert.Equal(t, "The request body has 1 error(s)", validationErr.Message)
		assert.Len(t, validationErr.Details, 1)
		assert.Equal(t, "Size must be between 1 and 5", *validationErr.Details[0].Message)
	})

	t.Run("returns error for invalid message type", func(t *testing.T) {
		request := messagingapi.ValidateBroadcastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{Type: "invalid"},
				},
			},
		}

		_, err := s.ValidateBroadcast(context.Background(), request)
		require.Error(t, err)
		
		var validationErr *ValidationError
		require.True(t, errors.As(err, &validationErr))
		assert.Equal(t, "The request body has 1 error(s)", validationErr.Message)
		assert.Contains(t, *validationErr.Details[0].Message, "Invalid message type")
	})

	t.Run("returns error for missing message type", func(t *testing.T) {
		request := messagingapi.ValidateBroadcastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{},
				},
			},
		}

		_, err := s.ValidateBroadcast(context.Background(), request)
		require.Error(t, err)
		
		var validationErr *ValidationError
		require.True(t, errors.As(err, &validationErr))
		assert.Equal(t, "The request body has 1 error(s)", validationErr.Message)
		assert.Equal(t, "Message type is required", *validationErr.Details[0].Message)
	})

	t.Run("returns error for nil request body", func(t *testing.T) {
		request := messagingapi.ValidateBroadcastRequestObject{
			Body: nil,
		}

		_, err := s.ValidateBroadcast(context.Background(), request)
		require.Error(t, err)
		
		var validationErr *ValidationError
		require.True(t, errors.As(err, &validationErr))
		assert.Equal(t, "Request body is required", validationErr.Message)
	})
}

func TestValidateMulticast(t *testing.T) {
	s := &server{}

	t.Run("validates valid multicast message", func(t *testing.T) {
		request := messagingapi.ValidateMulticastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{Type: "image"},
					{Type: "video"},
				},
			},
		}

		response, err := s.ValidateMulticast(context.Background(), request)
		require.NoError(t, err)
		assert.IsType(t, messagingapi.ValidateMulticast200Response{}, response)
	})
}

func TestValidateNarrowcast(t *testing.T) {
	s := &server{}

	t.Run("validates valid narrowcast message", func(t *testing.T) {
		request := messagingapi.ValidateNarrowcastRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{Type: "audio"},
					{Type: "file"},
					{Type: "location"},
				},
			},
		}

		response, err := s.ValidateNarrowcast(context.Background(), request)
		require.NoError(t, err)
		assert.IsType(t, messagingapi.ValidateNarrowcast200Response{}, response)
	})
}

func TestValidatePush(t *testing.T) {
	s := &server{}

	t.Run("validates valid push message", func(t *testing.T) {
		request := messagingapi.ValidatePushRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{Type: "sticker"},
					{Type: "template"},
				},
			},
		}

		response, err := s.ValidatePush(context.Background(), request)
		require.NoError(t, err)
		assert.IsType(t, messagingapi.ValidatePush200Response{}, response)
	})
}

func TestValidateReply(t *testing.T) {
	s := &server{}

	t.Run("validates valid reply message", func(t *testing.T) {
		request := messagingapi.ValidateReplyRequestObject{
			Body: &messagingapi.ValidateMessageRequest{
				Messages: []messagingapi.Message{
					{Type: "imagemap"},
					{Type: "flex"},
				},
			},
		}

		response, err := s.ValidateReply(context.Background(), request)
		require.NoError(t, err)
		assert.IsType(t, messagingapi.ValidateReply200Response{}, response)
	})

	t.Run("validates all supported message types", func(t *testing.T) {
		supportedTypes := []string{
			"text", "image", "video", "audio", "file",
			"location", "sticker", "template", "imagemap", "flex",
		}

		for _, msgType := range supportedTypes {
			request := messagingapi.ValidateReplyRequestObject{
				Body: &messagingapi.ValidateMessageRequest{
					Messages: []messagingapi.Message{
						{Type: msgType},
					},
				},
			}

			response, err := s.ValidateReply(context.Background(), request)
			require.NoError(t, err, "failed for message type: %s", msgType)
			assert.IsType(t, messagingapi.ValidateReply200Response{}, response)
		}
	})
}