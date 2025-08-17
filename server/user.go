package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// GetProfile gets user profile information
func (s *server) GetProfile(ctx context.Context, request messagingapi.GetProfileRequestObject) (messagingapi.GetProfileResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetFollowers gets follower IDs
func (s *server) GetFollowers(ctx context.Context, request messagingapi.GetFollowersRequestObject) (messagingapi.GetFollowersResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// IssueLinkToken issues a link token for account linking
func (s *server) IssueLinkToken(ctx context.Context, request messagingapi.IssueLinkTokenRequestObject) (messagingapi.IssueLinkTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}