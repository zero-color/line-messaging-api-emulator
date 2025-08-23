package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
	"github.com/zero-color/line-messaging-api-emulator/server"
)

func TestGetProfile(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)

	// Create a bot
	bot, err := dbClient.CreateBot(context.Background(), db.CreateBotParams{
		UserID:         "test-bot-id",
		BasicID:        "test-basic-id",
		ChatMode:       "chat",
		DisplayName:    "Test Bot",
		MarkAsReadMode: "auto",
	})
	require.NoError(t, err)

	// Create a user
	pictureUrl := "https://example.com/picture.jpg"
	statusMessage := "I love LINE!"
	language := "en"
	user, err := dbClient.CreateUser(context.Background(), db.CreateUserParams{
		UserID:        "test-user-id",
		DisplayName:   "Test User",
		PictureUrl:    &pictureUrl,
		StatusMessage: &statusMessage,
		Language:      &language,
	})
	require.NoError(t, err)

	t.Run("success - user is a follower", func(t *testing.T) {
		// Make the user a follower of the bot
		_, err := dbClient.CreateBotFollower(context.Background(), db.CreateBotFollowerParams{
			BotID:  bot.ID,
			UserID: user.ID,
		})
		require.NoError(t, err)

		// Set bot ID in context
		ctx := auth.SetBotID(context.Background(), bot.ID)

		resp, err := srv.GetProfile(ctx, messagingapi.GetProfileRequestObject{
			UserId: "test-user-id",
		})

		require.NoError(t, err)
		profile, ok := resp.(messagingapi.GetProfile200JSONResponse)
		assert.True(t, ok)
		assert.Equal(t, "test-user-id", profile.UserId)
		assert.Equal(t, "Test User", profile.DisplayName)
		assert.NotNil(t, profile.PictureUrl)
		assert.Equal(t, "https://example.com/picture.jpg", *profile.PictureUrl)
		assert.NotNil(t, profile.StatusMessage)
		assert.Equal(t, "I love LINE!", *profile.StatusMessage)
		assert.NotNil(t, profile.Language)
		assert.Equal(t, "en", *profile.Language)
	})

	t.Run("error - user is not a follower", func(t *testing.T) {
		// Create another bot
		otherBot, err := dbClient.CreateBot(context.Background(), db.CreateBotParams{
			UserID:         "other-bot-id",
			BasicID:        "other-basic-id",
			ChatMode:       "bot",
			DisplayName:    "Other Bot",
			MarkAsReadMode: "manual",
		})
		require.NoError(t, err)

		// Set the other bot ID in context
		ctx := auth.SetBotID(context.Background(), otherBot.ID)

		resp, err := srv.GetProfile(ctx, messagingapi.GetProfileRequestObject{
			UserId: "test-user-id",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("error - user does not exist", func(t *testing.T) {
		// Set bot ID in context
		ctx := auth.SetBotID(context.Background(), bot.ID)

		resp, err := srv.GetProfile(ctx, messagingapi.GetProfileRequestObject{
			UserId: "non-existent-user",
		})

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("success - user without optional fields", func(t *testing.T) {
		// Create a user without optional fields
		simpleUser, err := dbClient.CreateUser(context.Background(), db.CreateUserParams{
			UserID:      "simple-user-id",
			DisplayName: "Simple User",
		})
		require.NoError(t, err)

		// Make the user a follower of the bot
		_, err = dbClient.CreateBotFollower(context.Background(), db.CreateBotFollowerParams{
			BotID:  bot.ID,
			UserID: simpleUser.ID,
		})
		require.NoError(t, err)

		// Set bot ID in context
		ctx := auth.SetBotID(context.Background(), bot.ID)

		resp, err := srv.GetProfile(ctx, messagingapi.GetProfileRequestObject{
			UserId: "simple-user-id",
		})

		require.NoError(t, err)
		profile, ok := resp.(messagingapi.GetProfile200JSONResponse)
		assert.True(t, ok)
		assert.Equal(t, "simple-user-id", profile.UserId)
		assert.Equal(t, "Simple User", profile.DisplayName)
		assert.Nil(t, profile.PictureUrl)
		assert.Nil(t, profile.StatusMessage)
		assert.Nil(t, profile.Language)
	})
}