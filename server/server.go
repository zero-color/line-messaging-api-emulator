package server

import (
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/insightapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/webhook"
)

type Server interface {
	messagingapi.StrictServerInterface
	insightapi.StrictServerInterface
	adminapi.StrictServerInterface
}

type server struct {
	webhookClient webhook.Client
}

func New(webhookClient webhook.Client) Server {
	return &server{
		webhookClient: webhookClient,
	}
}
