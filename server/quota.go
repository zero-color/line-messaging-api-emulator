package server

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

// quotaTimeZone is the time zone in which the monthly message quota resets (UTC+9)
var quotaTimeZone = time.FixedZone("JST", 9*60*60)

// GetMessageQuota gets the monthly message quota. Bots in the emulator have no limit.
func (s *server) GetMessageQuota(ctx context.Context, request messagingapi.GetMessageQuotaRequestObject) (messagingapi.GetMessageQuotaResponseObject, error) {
	return messagingapi.GetMessageQuota200JSONResponse{Type: messagingapi.None}, nil
}

// GetMessageQuotaConsumption gets the number of messages sent this month that count towards the quota
func (s *server) GetMessageQuotaConsumption(ctx context.Context, request messagingapi.GetMessageQuotaConsumptionRequestObject) (messagingapi.GetMessageQuotaConsumptionResponseObject, error) {
	now := time.Now().In(quotaTimeZone)
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, quotaTimeZone)

	totalUsage, err := db.FromContext(ctx).Queries().CountMessagesSentSince(ctx, db.CountMessagesSentSinceParams{
		BotID: auth.GetBotID(ctx),
		Since: pgtype.Timestamptz{Time: startOfMonth, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count sent messages: %w", err)
	}
	return messagingapi.GetMessageQuotaConsumption200JSONResponse{TotalUsage: totalUsage}, nil
}
