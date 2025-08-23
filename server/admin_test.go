package server_test

import (
	"context"
	"strings"
	"testing"

	"github.com/samber/lo"

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

func TestCreateFollowers(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)
	ctx := context.Background()

	// First, create a bot
	createBotReq := adminapi.CreateBotRequestObject{
		Body: &adminapi.CreateBotRequest{
			DisplayName: "Test Bot for Followers",
			BasicId:     lo.ToPtr("@testfollowers"),
		},
	}

	botResp, err := srv.CreateBot(ctx, createBotReq)
	require.NoError(t, err)
	botInfo, ok := botResp.(adminapi.CreateBot201JSONResponse)
	require.True(t, ok)

	t.Run("create multiple followers successfully", func(t *testing.T) {
		req := adminapi.CreateFollowersRequestObject{
			BotId: botInfo.UserId,
			Body: &adminapi.CreateFollowersRequest{
				Count: 5,
			},
		}

		resp, err := srv.CreateFollowers(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(adminapi.CreateFollowers201JSONResponse)
		require.True(t, ok)
		assert.Equal(t, 5, successResp.Count)
		assert.Len(t, successResp.Followers, 5)

		// Verify all followers have required fields
		for _, follower := range successResp.Followers {
			assert.NotEmpty(t, follower.UserId)
			assert.NotEmpty(t, follower.DisplayName)
			assert.True(t, strings.HasPrefix(follower.UserId, "U"))
		}
	})

	t.Run("create followers for non-existent bot", func(t *testing.T) {
		req := adminapi.CreateFollowersRequestObject{
			BotId: "U_nonexistent_bot_id",
			Body: &adminapi.CreateFollowersRequest{
				Count: 3,
			},
		}

		resp, err := srv.CreateFollowers(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(adminapi.CreateFollowers404JSONResponse)
		assert.True(t, ok, "Expected 404 response for non-existent bot")
	})

	t.Run("invalid count - too low", func(t *testing.T) {
		req := adminapi.CreateFollowersRequestObject{
			BotId: botInfo.UserId,
			Body: &adminapi.CreateFollowersRequest{
				Count: 0,
			},
		}

		resp, err := srv.CreateFollowers(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(adminapi.CreateFollowers400JSONResponse)
		assert.True(t, ok, "Expected 400 response for invalid count")
	})

	t.Run("invalid count - too high", func(t *testing.T) {
		req := adminapi.CreateFollowersRequestObject{
			BotId: botInfo.UserId,
			Body: &adminapi.CreateFollowersRequest{
				Count: 1001,
			},
		}

		resp, err := srv.CreateFollowers(ctx, req)
		require.NoError(t, err)

		_, ok := resp.(adminapi.CreateFollowers400JSONResponse)
		assert.True(t, ok, "Expected 400 response for count exceeding limit")
	})

	t.Run("followers have varied optional fields", func(t *testing.T) {
		req := adminapi.CreateFollowersRequestObject{
			BotId: botInfo.UserId,
			Body: &adminapi.CreateFollowersRequest{
				Count: 20, // Create more to ensure variety
			},
		}

		resp, err := srv.CreateFollowers(ctx, req)
		require.NoError(t, err)

		successResp, ok := resp.(adminapi.CreateFollowers201JSONResponse)
		require.True(t, ok)

		// Check that at least some followers have optional fields
		hasLanguage := false
		hasPictureUrl := false
		hasStatusMessage := false
		hasEmptyStatus := false

		for _, follower := range successResp.Followers {
			if follower.Language != nil && *follower.Language != "" {
				hasLanguage = true
			}
			if follower.PictureUrl != nil && *follower.PictureUrl != "" {
				hasPictureUrl = true
			}
			if follower.StatusMessage != nil && *follower.StatusMessage != "" {
				hasStatusMessage = true
			}
			if follower.StatusMessage == nil || *follower.StatusMessage == "" {
				hasEmptyStatus = true
			}
		}

		// With 20 followers and the probability settings,
		// we should have variety in optional fields
		assert.True(t, hasLanguage, "Expected some followers to have language")
		assert.True(t, hasPictureUrl, "Expected some followers to have picture URL")
		assert.True(t, hasStatusMessage || hasEmptyStatus, "Expected variety in status messages")
	})
}
