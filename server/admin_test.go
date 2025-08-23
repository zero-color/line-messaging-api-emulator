package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/server"
)

func TestCreateBot(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)

	ctx := context.Background()

	t.Run("create bot with minimal fields", func(t *testing.T) {
		displayName := "Test Bot"
		req := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: displayName,
			},
		}

		resp, err := srv.CreateBot(ctx, req)
		require.NoError(t, err)

		botResp, ok := resp.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, displayName, botResp.DisplayName)
		assert.NotEmpty(t, botResp.UserId)
		assert.Equal(t, adminapi.BotInfoResponseChatModeBot, botResp.ChatMode)
		assert.Equal(t, adminapi.BotInfoResponseMarkAsReadModeManual, botResp.MarkAsReadMode)
	})

	t.Run("create bot with all fields", func(t *testing.T) {
		userID := "U123456789"
		basicID := "@testbot"
		displayName := "Full Test Bot"
		pictureURL := "https://example.com/picture.jpg"
		premiumID := "premium123"
		chatMode := adminapi.CreateBotRequestChatModeChat
		markAsReadMode := adminapi.CreateBotRequestMarkAsReadModeAuto

		req := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				UserId:         &userID,
				BasicId:        &basicID,
				DisplayName:    displayName,
				PictureUrl:     &pictureURL,
				PremiumId:      &premiumID,
				ChatMode:       &chatMode,
				MarkAsReadMode: &markAsReadMode,
			},
		}

		resp, err := srv.CreateBot(ctx, req)
		require.NoError(t, err)

		botResp, ok := resp.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, userID, botResp.UserId)
		assert.Equal(t, basicID, botResp.BasicId)
		assert.Equal(t, displayName, botResp.DisplayName)
		assert.Equal(t, &pictureURL, botResp.PictureUrl)
		assert.Equal(t, &premiumID, botResp.PremiumId)
		assert.Equal(t, adminapi.BotInfoResponseChatModeChat, botResp.ChatMode)
		assert.Equal(t, adminapi.BotInfoResponseMarkAsReadModeAuto, botResp.MarkAsReadMode)
	})

	t.Run("create duplicate bot returns conflict", func(t *testing.T) {
		userID := "U_duplicate"
		displayName := "Duplicate Bot"

		req := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				UserId:      &userID,
				DisplayName: displayName,
			},
		}

		resp1, err := srv.CreateBot(ctx, req)
		require.NoError(t, err)
		_, ok := resp1.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		resp2, err := srv.CreateBot(ctx, req)
		require.NoError(t, err)
		conflictResp, ok := resp2.(adminapi.CreateBot409JSONResponse)
		require.True(t, ok, "Expected CreateBot409JSONResponse, got %T", resp2)
		assert.Contains(t, conflictResp.Error.Message, "already exists")
	})
}
