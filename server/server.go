package server

import (
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

type Server interface {
	messagingapi.ServerInterface
	adminapi.ServerInterface
}

type server struct {
}

var _ Server = (*server)(nil)

func New() Server {
	return &server{}
}
