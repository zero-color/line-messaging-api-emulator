package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
	"github.com/zero-color/line-messaging-api-emulator/internal/xuuid"
	"strings"
)

// handleRetryKey processes the retry key from request parameters and checks for duplicate messages
// Returns the parsed retry key UUID and true if the message already exists (should return early)
func (s *server) handleRetryKey(ctx context.Context, retryKey *uuid.UUID) (pgtype.UUID, bool, error) {
	var retryKeyUUID pgtype.UUID

	if retryKey == nil {
		return retryKeyUUID, false, nil
	}

	retryKeyStr := retryKey.String()
	parsedUUID, err := uuid.Parse(retryKeyStr)
	if err != nil {
		// Invalid UUID format, just return empty UUID
		return retryKeyUUID, false, nil
	}

	retryKeyUUID.Bytes = parsedUUID
	retryKeyUUID.Valid = true

	// Check if message with this retry key already exists
	_, err = s.db.GetMessagesByRetryKey(ctx, retryKeyUUID)
	if err == nil {
		// Message already exists, indicate early return
		return retryKeyUUID, true, nil
	}

	// Message doesn't exist, proceed with normal flow
	return retryKeyUUID, false, nil
}

// Broadcast sends a message to all users
func (s *server) Broadcast(ctx context.Context, request messagingapi.BroadcastRequestObject) (messagingapi.BroadcastResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	// Validate messages
	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	botID := auth.GetBotID(ctx)

	// Handle retry key for idempotency
	retryKeyUUID, isDuplicate, err := s.handleRetryKey(ctx, request.Params.XLineRetryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to handle retry key: %w", err)
	}
	if isDuplicate {
		return messagingapi.Broadcast200JSONResponse{}, nil
	}

	// Serialize messages to JSON
	messagesJSON, err := json.Marshal(request.Body.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize messages: %w", err)
	}

	// Store the message in database
	recipientType := "all"

	_, err = s.db.CreateMessage(ctx, db.CreateMessageParams{
		BotID:         botID,
		MessageType:   "broadcast",
		RecipientType: &recipientType,
		RecipientID:   nil,
		Content:       messagesJSON,
		RetryKey:      retryKeyUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	// Return empty response on success
	return messagingapi.Broadcast200JSONResponse{}, nil
}

// Multicast sends a message to multiple users
func (s *server) Multicast(ctx context.Context, request messagingapi.MulticastRequestObject) (messagingapi.MulticastResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	// Validate messages
	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	// Validate recipient list
	if len(request.Body.To) == 0 {
		validationErr := NewValidationError("The request body has 1 error(s)")
		validationErr.AddDetail("Size must be between 1 and 500", "to")
		return nil, validationErr
	}

	if len(request.Body.To) > 500 {
		validationErr := NewValidationError("The request body has 1 error(s)")
		validationErr.AddDetail("Size must be between 1 and 500", "to")
		return nil, validationErr
	}

	botID := auth.GetBotID(ctx)

	// Handle retry key for idempotency
	retryKeyUUID, isDuplicate, err := s.handleRetryKey(ctx, request.Params.XLineRetryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to handle retry key: %w", err)
	}
	if isDuplicate {
		return messagingapi.Multicast200JSONResponse{}, nil
	}

	// Serialize messages to JSON
	messagesJSON, err := json.Marshal(request.Body.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize messages: %w", err)
	}

	// Store recipient IDs as comma-separated string
	recipientIDs := strings.Join(request.Body.To, ",")
	recipientType := "multiple"

	_, err = s.db.CreateMessage(ctx, db.CreateMessageParams{
		BotID:         botID,
		MessageType:   "multicast",
		RecipientType: &recipientType,
		RecipientID:   &recipientIDs,
		Content:       messagesJSON,
		RetryKey:      retryKeyUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	// Return empty response on success
	return messagingapi.Multicast200JSONResponse{}, nil
}

// Narrowcast sends a message to a narrowed audience
func (s *server) Narrowcast(ctx context.Context, request messagingapi.NarrowcastRequestObject) (messagingapi.NarrowcastResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	// Validate messages
	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	botID := auth.GetBotID(ctx)

	// Handle retry key for idempotency
	retryKeyUUID, isDuplicate, err := s.handleRetryKey(ctx, request.Params.XLineRetryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to handle retry key: %w", err)
	}
	if isDuplicate {
		return messagingapi.Narrowcast202JSONResponse{}, nil
	}

	// Serialize messages to JSON
	messagesJSON, err := json.Marshal(request.Body.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize messages: %w", err)
	}

	// Serialize filter if provided
	var filterJSON []byte
	if request.Body.Filter != nil {
		filterJSON, err = json.Marshal(request.Body.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize filter: %w", err)
		}
	}

	// Store the message in database
	recipientType := "filtered"
	var recipientID *string
	if filterJSON != nil {
		filterStr := string(filterJSON)
		recipientID = &filterStr
	}

	_, err = s.db.CreateMessage(ctx, db.CreateMessageParams{
		BotID:         botID,
		MessageType:   "narrowcast",
		RecipientType: &recipientType,
		RecipientID:   recipientID,
		Content:       messagesJSON,
		RetryKey:      retryKeyUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	// Return empty response on success
	return messagingapi.Narrowcast202JSONResponse{}, nil
}

// GetNarrowcastProgress gets the progress of a narrowcast message
func (s *server) GetNarrowcastProgress(ctx context.Context, request messagingapi.GetNarrowcastProgressRequestObject) (messagingapi.GetNarrowcastProgressResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// PushMessage sends a push message to a single user
func (s *server) PushMessage(ctx context.Context, request messagingapi.PushMessageRequestObject) (messagingapi.PushMessageResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	// Validate messages
	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	botID := auth.GetBotID(ctx)

	// Handle retry key for idempotency
	retryKeyUUID, isDuplicate, err := s.handleRetryKey(ctx, request.Params.XLineRetryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to handle retry key: %w", err)
	}
	if isDuplicate {
		return messagingapi.PushMessage200JSONResponse{
			SentMessages: []messagingapi.SentMessage{},
		}, nil
	}

	// Serialize messages to JSON
	messagesJSON, err := json.Marshal(request.Body.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize messages: %w", err)
	}

	// Store the message in database
	recipientType := "user"
	recipientID := request.Body.To

	msg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		BotID:         botID,
		MessageType:   "push",
		RecipientType: &recipientType,
		RecipientID:   &recipientID,
		Content:       messagesJSON,
		RetryKey:      retryKeyUUID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	var sentMessages []messagingapi.SentMessage
	for i := range request.Body.Messages {
		sentMessages = append(sentMessages, messagingapi.SentMessage{
			Id:         fmt.Sprintf("%d_%d", msg.ID, i+1),
			QuoteToken: lo.ToPtr(fmt.Sprintf("%d_%d", msg.ID, i+1)),
		})
	}

	return messagingapi.PushMessage200JSONResponse{
		SentMessages: sentMessages,
	}, nil
}

// PushMessagesByPhone sends push messages by phone number
func (s *server) PushMessagesByPhone(ctx context.Context, request messagingapi.PushMessagesByPhoneRequestObject) (messagingapi.PushMessagesByPhoneResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ReplyMessage sends a reply message
func (s *server) ReplyMessage(ctx context.Context, request messagingapi.ReplyMessageRequestObject) (messagingapi.ReplyMessageResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	// Validate messages
	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	botID := auth.GetBotID(ctx)

	// We store the reply token as a UUID in the RetryKey column for uniqueness checking
	parsedUUID := uuid.NewMD5(xuuid.NameSpaceReplyToken, []byte(request.Body.ReplyToken))
	replyTokenUUID := pgtype.UUID{
		Bytes: parsedUUID,
		Valid: true,
	}

	// Check if this reply token has been used before
	if _, err := s.db.GetMessagesByRetryKey(ctx, replyTokenUUID); err == nil {
		// Reply token has already been used, return 400 error
		validationErr := NewValidationError("Invalid reply token")
		validationErr.AddDetail("Reply token has already been used or expired", "replyToken")
		return nil, validationErr
	}

	// Serialize messages to JSON
	messagesJSON, err := json.Marshal(request.Body.Messages)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize messages: %w", err)
	}

	// Store the message in database
	messageType := "reply"
	recipientType := "reply"

	msg, err := s.db.CreateMessage(ctx, db.CreateMessageParams{
		BotID:         botID,
		MessageType:   messageType,
		RecipientType: &recipientType,
		RecipientID:   &request.Body.ReplyToken,
		Content:       messagesJSON,
		RetryKey:      replyTokenUUID, // Store reply token UUID for duplicate checking
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	// Generate sent messages with IDs for each message
	var sentMessages []messagingapi.SentMessage
	for i := range request.Body.Messages {
		sentMessages = append(sentMessages, messagingapi.SentMessage{
			Id:         fmt.Sprintf("%d_%d", msg.ID, i+1),
			QuoteToken: lo.ToPtr(fmt.Sprintf("%d_%d", msg.ID, i+1)),
		})
	}

	// Return response with sent messages
	return messagingapi.ReplyMessage200JSONResponse{
		SentMessages: sentMessages,
	}, nil
}

// ValidateBroadcast validates a broadcast message
func (s *server) ValidateBroadcast(ctx context.Context, request messagingapi.ValidateBroadcastRequestObject) (messagingapi.ValidateBroadcastResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	return messagingapi.ValidateBroadcast200Response{}, nil
}

// ValidateMulticast validates a multicast message
func (s *server) ValidateMulticast(ctx context.Context, request messagingapi.ValidateMulticastRequestObject) (messagingapi.ValidateMulticastResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	return messagingapi.ValidateMulticast200Response{}, nil
}

// ValidateNarrowcast validates a narrowcast message
func (s *server) ValidateNarrowcast(ctx context.Context, request messagingapi.ValidateNarrowcastRequestObject) (messagingapi.ValidateNarrowcastResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	return messagingapi.ValidateNarrowcast200Response{}, nil
}

// ValidatePush validates a push message
func (s *server) ValidatePush(ctx context.Context, request messagingapi.ValidatePushRequestObject) (messagingapi.ValidatePushResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	return messagingapi.ValidatePush200Response{}, nil
}

// ValidateReply validates a reply message
func (s *server) ValidateReply(ctx context.Context, request messagingapi.ValidateReplyRequestObject) (messagingapi.ValidateReplyResponseObject, error) {
	if request.Body == nil {
		return nil, NewValidationError("Request body is required")
	}

	if err := validateMessages(request.Body.Messages); err != nil {
		return nil, err
	}

	return messagingapi.ValidateReply200Response{}, nil
}

// GetMessageContent gets the content of a message
func (s *server) GetMessageContent(ctx context.Context, request messagingapi.GetMessageContentRequestObject) (messagingapi.GetMessageContentResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetMessageContentPreview gets the preview of message content
func (s *server) GetMessageContentPreview(ctx context.Context, request messagingapi.GetMessageContentPreviewRequestObject) (messagingapi.GetMessageContentPreviewResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetMessageContentTranscodingByMessageId gets transcoding status by message ID
func (s *server) GetMessageContentTranscodingByMessageId(ctx context.Context, request messagingapi.GetMessageContentTranscodingByMessageIdRequestObject) (messagingapi.GetMessageContentTranscodingByMessageIdResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// MarkMessagesAsRead marks messages as read
func (s *server) MarkMessagesAsRead(ctx context.Context, request messagingapi.MarkMessagesAsReadRequestObject) (messagingapi.MarkMessagesAsReadResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ShowLoadingAnimation shows loading animation
func (s *server) ShowLoadingAnimation(ctx context.Context, request messagingapi.ShowLoadingAnimationRequestObject) (messagingapi.ShowLoadingAnimationResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// validateMessages validates an array of message objects
func validateMessages(messages []messagingapi.Message) error {
	if len(messages) == 0 {
		validationErr := NewValidationError("The request body has 1 error(s)")
		validationErr.AddDetail("Size must be between 1 and 5", "messages")
		return validationErr
	}

	if len(messages) > 5 {
		validationErr := NewValidationError("The request body has 1 error(s)")
		validationErr.AddDetail("Size must be between 1 and 5", "messages")
		return validationErr
	}

	// Collect all validation errors
	var errorCount int
	validationErr := NewValidationError("")

	for i, msg := range messages {
		if err := validateMessage(msg, i, validationErr); err != nil {
			errorCount++
		}
	}

	if errorCount > 0 {
		validationErr.Message = fmt.Sprintf("The request body has %d error(s)", errorCount)
		return validationErr
	}

	return nil
}

// validateMessage validates a single message object
func validateMessage(msg messagingapi.Message, index int, validationErr *ValidationError) error {
	var hasError bool

	if msg.Type == "" {
		validationErr.AddDetail("Message type is required", fmt.Sprintf("messages[%d].type", index))
		hasError = true
	} else {
		// Validate based on message type
		switch msg.Type {
		case "text", "image", "video", "audio", "file", "location", "sticker", "template", "imagemap", "flex":
			// These are valid message types
			// TODO: Add specific validation for each message type
		default:
			validationErr.AddDetail(fmt.Sprintf("Invalid message type: %s", msg.Type), fmt.Sprintf("messages[%d].type", index))
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("validation failed")
	}
	return nil
}
