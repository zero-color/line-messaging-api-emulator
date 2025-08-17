package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// GetWebhookEndpoint gets the webhook endpoint URL
func (s *server) GetWebhookEndpoint(ctx context.Context, request messagingapi.GetWebhookEndpointRequestObject) (messagingapi.GetWebhookEndpointResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// SetWebhookEndpoint sets the webhook endpoint URL
func (s *server) SetWebhookEndpoint(ctx context.Context, request messagingapi.SetWebhookEndpointRequestObject) (messagingapi.SetWebhookEndpointResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// TestWebhookEndpoint tests the webhook endpoint
func (s *server) TestWebhookEndpoint(ctx context.Context, request messagingapi.TestWebhookEndpointRequestObject) (messagingapi.TestWebhookEndpointResponseObject, error) {
	//TODO implement me
	panic("implement me")
}