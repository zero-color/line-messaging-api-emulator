package server

import (
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

type Server interface {
	messagingapi.StrictServerInterface
	adminapi.StrictServerInterface
}

type server struct {
}

var _ Server = (*server)(nil)

func New() Server {
	return &server{}
}
