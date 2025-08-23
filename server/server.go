package server

import (
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
)

type Server interface {
	messagingapi.StrictServerInterface
	adminapi.StrictServerInterface
}

type server struct {
	db db.Querier
}

func New(db db.Querier) Server {
	return &server{
		db: db,
	}
}
