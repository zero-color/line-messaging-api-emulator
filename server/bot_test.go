package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
	"github.com/zero-color/line-messaging-api-emulator/server"
)

func TestGetBotInfo(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)

	ctx := context.Background()

	t.Run("returns bot info when bot exists", func(t *testing.T) {
		// Create a bot first
		displayName := "Test Bot"
		pictureURL := "https://example.com/pic.jpg"
		premiumID := "premium123"
		createReq := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: displayName,
				PictureUrl:  &pictureURL,
				PremiumId:   &premiumID,
			},
		}

		createResp, err := srv.CreateBot(ctx, createReq)
		require.NoError(t, err)
		createdBot, ok := createResp.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		// Get the bot from database to get its ID
		bot, err := dbClient.GetBotByUserID(ctx, createdBot.UserId)
		require.NoError(t, err)

		// Add bot ID to context
		ctx := auth.SetBotID(context.Background(), bot.ID)

		// Get bot info
		req := messagingapi.GetBotInfoRequestObject{}
		resp, err := srv.GetBotInfo(ctx, req)
		require.NoError(t, err)

		botResp, ok := resp.(messagingapi.GetBotInfo200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, createdBot.UserId, botResp.UserId)
		assert.Equal(t, createdBot.BasicId, botResp.BasicId)
		assert.Equal(t, displayName, botResp.DisplayName)
		assert.Equal(t, &pictureURL, botResp.PictureUrl)
		assert.Equal(t, &premiumID, botResp.PremiumId)
		assert.Equal(t, messagingapi.BotInfoResponseChatMode("bot"), botResp.ChatMode)
		assert.Equal(t, messagingapi.BotInfoResponseMarkAsReadMode("manual"), botResp.MarkAsReadMode)
	})

	t.Run("returns error when bot does not exist", func(t *testing.T) {
		// Use a non-existent bot ID
		ctx := auth.SetBotID(context.Background(), 99999)

		req := messagingapi.GetBotInfoRequestObject{}
		resp, err := srv.GetBotInfo(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get bot")
		assert.Nil(t, resp)
	})

	t.Run("returns correct bot for specific bot ID", func(t *testing.T) {
		// Create first bot
		firstBotName := "First Bot"
		createReq1 := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: firstBotName,
			},
		}
		createResp1, err := srv.CreateBot(ctx, createReq1)
		require.NoError(t, err)
		firstBot, ok := createResp1.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		// Create second bot
		secondBotName := "Second Bot"
		createReq2 := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: secondBotName,
			},
		}
		createResp2, err := srv.CreateBot(ctx, createReq2)
		require.NoError(t, err)
		secondBot, ok := createResp2.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		// Get first bot from database to get its ID
		bot1, err := dbClient.GetBotByUserID(ctx, firstBot.UserId)
		require.NoError(t, err)

		// Get second bot from database to get its ID
		bot2, err := dbClient.GetBotByUserID(ctx, secondBot.UserId)
		require.NoError(t, err)

		// Test with first bot ID
		ctxWithBot1 := auth.SetBotID(ctx, bot1.ID)
		req := messagingapi.GetBotInfoRequestObject{}
		resp1, err := srv.GetBotInfo(ctxWithBot1, req)
		require.NoError(t, err)
		botResp1, ok := resp1.(messagingapi.GetBotInfo200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, firstBot.UserId, botResp1.UserId)
		assert.Equal(t, firstBotName, botResp1.DisplayName)

		// Test with second bot ID
		ctxWithBot2 := auth.SetBotID(ctx, bot2.ID)
		resp2, err := srv.GetBotInfo(ctxWithBot2, req)
		require.NoError(t, err)
		botResp2, ok := resp2.(messagingapi.GetBotInfo200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, secondBot.UserId, botResp2.UserId)
		assert.Equal(t, secondBotName, botResp2.DisplayName)
	})
}
