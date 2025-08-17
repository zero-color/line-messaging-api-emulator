package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api"
)

func (s server) LeaveGroup(w http.ResponseWriter, r *http.Request, groupId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupMemberProfile(w http.ResponseWriter, r *http.Request, groupId string, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupMemberCount(w http.ResponseWriter, r *http.Request, groupId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupMembersIds(w http.ResponseWriter, r *http.Request, groupId string, params api.GetGroupMembersIdsParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetGroupSummary(w http.ResponseWriter, r *http.Request, groupId string) {
	//TODO implement me
	panic("implement me")
}
