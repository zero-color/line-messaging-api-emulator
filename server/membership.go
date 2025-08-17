package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api"
)

func (s server) GetMembershipList(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetMembershipSubscription(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetJoinedMembershipUsers(w http.ResponseWriter, r *http.Request, membershipId int, params api.GetJoinedMembershipUsersParams) {
	//TODO implement me
	panic("implement me")
}
