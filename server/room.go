package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// LeaveRoom leaves a room
func (s *server) LeaveRoom(ctx context.Context, request messagingapi.LeaveRoomRequestObject) (messagingapi.LeaveRoomResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetRoomMemberCount gets the member count of a room
func (s *server) GetRoomMemberCount(ctx context.Context, request messagingapi.GetRoomMemberCountRequestObject) (messagingapi.GetRoomMemberCountResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetRoomMemberProfile gets the profile of a room member
func (s *server) GetRoomMemberProfile(ctx context.Context, request messagingapi.GetRoomMemberProfileRequestObject) (messagingapi.GetRoomMemberProfileResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetRoomMembersIds gets the user IDs of room members
func (s *server) GetRoomMembersIds(ctx context.Context, request messagingapi.GetRoomMembersIdsRequestObject) (messagingapi.GetRoomMembersIdsResponseObject, error) {
	//TODO implement me
	panic("implement me")
}