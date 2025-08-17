package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

func (s server) GetBotInfo(w http.ResponseWriter, r *http.Request) {
	messagingapi.BotInfoResponse{}
	//TODO implement me
	panic("implement me")
}
