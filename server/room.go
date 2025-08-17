package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

func (s *server) LeaveRoom(w http.ResponseWriter, r *http.Request, roomId string) {
	//TODO implement me
	panic("implement me")
}

func (s *server) GetRoomMemberProfile(w http.ResponseWriter, r *http.Request, roomId string, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s *server) GetRoomMemberCount(w http.ResponseWriter, r *http.Request, roomId string) {
	//TODO implement me
	panic("implement me")
}

func (s *server) GetRoomMembersIds(w http.ResponseWriter, r *http.Request, roomId string, params messagingapi.GetRoomMembersIdsParams) {
	//TODO implement me
	panic("implement me")
}
