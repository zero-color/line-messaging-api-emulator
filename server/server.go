package server

import (
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api"
)

type server struct {
}

var _ api.ServerInterface = (*server)(nil)

func New() api.ServerInterface {
	return &server{}
}

func (s server) GetBotInfo(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}
