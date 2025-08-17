package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// LeaveGroup leaves a group chat
func (s *server) LeaveGroup(ctx context.Context, request messagingapi.LeaveGroupRequestObject) (messagingapi.LeaveGroupResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetGroupMemberProfile gets the profile of a group member
func (s *server) GetGroupMemberProfile(ctx context.Context, request messagingapi.GetGroupMemberProfileRequestObject) (messagingapi.GetGroupMemberProfileResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetGroupMemberCount gets the member count of a group
func (s *server) GetGroupMemberCount(ctx context.Context, request messagingapi.GetGroupMemberCountRequestObject) (messagingapi.GetGroupMemberCountResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetGroupMembersIds gets the user IDs of group members
func (s *server) GetGroupMembersIds(ctx context.Context, request messagingapi.GetGroupMembersIdsRequestObject) (messagingapi.GetGroupMembersIdsResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetGroupSummary gets the group summary
func (s *server) GetGroupSummary(ctx context.Context, request messagingapi.GetGroupSummaryRequestObject) (messagingapi.GetGroupSummaryResponseObject, error) {
	//TODO implement me
	panic("implement me")
}