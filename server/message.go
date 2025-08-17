package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// Broadcast sends a message to all users
func (s *server) Broadcast(ctx context.Context, request messagingapi.BroadcastRequestObject) (messagingapi.BroadcastResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// Multicast sends a message to multiple users
func (s *server) Multicast(ctx context.Context, request messagingapi.MulticastRequestObject) (messagingapi.MulticastResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// Narrowcast sends a message to a narrowed audience
func (s *server) Narrowcast(ctx context.Context, request messagingapi.NarrowcastRequestObject) (messagingapi.NarrowcastResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetNarrowcastProgress gets the progress of a narrowcast message
func (s *server) GetNarrowcastProgress(ctx context.Context, request messagingapi.GetNarrowcastProgressRequestObject) (messagingapi.GetNarrowcastProgressResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// PushMessage sends a push message to a single user
func (s *server) PushMessage(ctx context.Context, request messagingapi.PushMessageRequestObject) (messagingapi.PushMessageResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// PushMessagesByPhone sends push messages by phone number
func (s *server) PushMessagesByPhone(ctx context.Context, request messagingapi.PushMessagesByPhoneRequestObject) (messagingapi.PushMessagesByPhoneResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ReplyMessage sends a reply message
func (s *server) ReplyMessage(ctx context.Context, request messagingapi.ReplyMessageRequestObject) (messagingapi.ReplyMessageResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ValidateBroadcast validates a broadcast message
func (s *server) ValidateBroadcast(ctx context.Context, request messagingapi.ValidateBroadcastRequestObject) (messagingapi.ValidateBroadcastResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ValidateMulticast validates a multicast message
func (s *server) ValidateMulticast(ctx context.Context, request messagingapi.ValidateMulticastRequestObject) (messagingapi.ValidateMulticastResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ValidateNarrowcast validates a narrowcast message
func (s *server) ValidateNarrowcast(ctx context.Context, request messagingapi.ValidateNarrowcastRequestObject) (messagingapi.ValidateNarrowcastResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ValidatePush validates a push message
func (s *server) ValidatePush(ctx context.Context, request messagingapi.ValidatePushRequestObject) (messagingapi.ValidatePushResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// ValidateReply validates a reply message
func (s *server) ValidateReply(ctx context.Context, request messagingapi.ValidateReplyRequestObject) (messagingapi.ValidateReplyResponseObject, error) {
	//TODO implement me
	panic("implement me")
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