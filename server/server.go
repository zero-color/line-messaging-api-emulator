package server

import (
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

type server struct {
}

var _ messagingapi.ServerInterface = (*server)(nil)

func New() messagingapi.ServerInterface {
	return &server{}
}
