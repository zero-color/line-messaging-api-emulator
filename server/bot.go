package server

import (
	"context"
	"fmt"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

// GetBotInfo gets information about the bot
func (s *server) GetBotInfo(ctx context.Context, _ messagingapi.GetBotInfoRequestObject) (messagingapi.GetBotInfoResponseObject, error) {
	bot, err := s.db.GetBot(ctx, auth.GetBotID(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get bot: %w", err)
	}

	response := messagingapi.GetBotInfo200JSONResponse{
		UserId:         bot.UserID,
		BasicId:        bot.BasicID,
		DisplayName:    bot.DisplayName,
		ChatMode:       messagingapi.BotInfoResponseChatMode(bot.ChatMode),
		MarkAsReadMode: messagingapi.BotInfoResponseMarkAsReadMode(bot.MarkAsReadMode),
	}

	if bot.PictureUrl.Valid {
		response.PictureUrl = &bot.PictureUrl.String
	}

	if bot.PremiumID.Valid {
		response.PremiumId = &bot.PremiumID.String
	}

	return response, nil
}
