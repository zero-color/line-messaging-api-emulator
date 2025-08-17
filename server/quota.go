package server

import (
	"context"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

// GetMessageQuota gets the message quota
func (s *server) GetMessageQuota(ctx context.Context, request messagingapi.GetMessageQuotaRequestObject) (messagingapi.GetMessageQuotaResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

// GetMessageQuotaConsumption gets the message quota consumption
func (s *server) GetMessageQuotaConsumption(ctx context.Context, request messagingapi.GetMessageQuotaConsumptionRequestObject) (messagingapi.GetMessageQuotaConsumptionResponseObject, error) {
	//TODO implement me
	panic("implement me")
}