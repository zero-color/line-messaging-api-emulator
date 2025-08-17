package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api"
)

func (s server) GetFollowers(w http.ResponseWriter, r *http.Request, params api.GetFollowersParams) {
	//TODO implement me
	panic("implement me")
}

func (s server) GetProfile(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}

func (s server) IssueLinkToken(w http.ResponseWriter, r *http.Request, userId string) {
	//TODO implement me
	panic("implement me")
}
