package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// GetMembershipList gets the list of memberships
func (s *server) GetMembershipList(ctx context.Context, request messagingapi.GetMembershipListRequestObject) (messagingapi.GetMembershipListResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetMembershipSubscription gets membership subscription information
func (s *server) GetMembershipSubscription(ctx context.Context, request messagingapi.GetMembershipSubscriptionRequestObject) (messagingapi.GetMembershipSubscriptionResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetJoinedMembershipUsers gets users who joined a membership
func (s *server) GetJoinedMembershipUsers(ctx context.Context, request messagingapi.GetJoinedMembershipUsersRequestObject) (messagingapi.GetJoinedMembershipUsersResponseObject, error) {
	//TODO implement me
	panic("implement me")
}