package server_test

import (
	"context"
	"fmt"
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

func TestGetFollowers(t *testing.T) {
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

	// Create multiple users
	userIDs := []string{}
	for i := 0; i < 5; i++ {
		userID := fmt.Sprintf("user-%d", i)
		user, err := dbClient.CreateUser(context.Background(), db.CreateUserParams{
			UserID:      userID,
			DisplayName: fmt.Sprintf("User %d", i),
		})
		require.NoError(t, err)

		// Make the user a follower of the bot
		_, err = dbClient.CreateBotFollower(context.Background(), db.CreateBotFollowerParams{
			BotID:  bot.ID,
			UserID: user.ID,
		})
		require.NoError(t, err)

		userIDs = append(userIDs, userID)
	}

	// Set bot ID in context
	ctx := auth.SetBotID(context.Background(), bot.ID)

	t.Run("success - get all followers", func(t *testing.T) {
		resp, err := srv.GetFollowers(ctx, messagingapi.GetFollowersRequestObject{})
		require.NoError(t, err)

		followers, ok := resp.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, followers.UserIds, 5)
		assert.Nil(t, followers.Next)

		// Check that all user IDs are present (order may differ due to followed_at DESC)
		for _, expectedID := range userIDs {
			assert.Contains(t, followers.UserIds, expectedID)
		}
	})

	t.Run("success - get followers with limit", func(t *testing.T) {
		limit := int32(2)
		resp, err := srv.GetFollowers(ctx, messagingapi.GetFollowersRequestObject{
			Params: messagingapi.GetFollowersParams{
				Limit: &limit,
			},
		})
		require.NoError(t, err)

		followers, ok := resp.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, followers.UserIds, 2)
		assert.NotNil(t, followers.Next)
		assert.Equal(t, "2", *followers.Next)
	})

	t.Run("success - get followers with pagination", func(t *testing.T) {
		// First page
		limit := int32(2)
		resp1, err := srv.GetFollowers(ctx, messagingapi.GetFollowersRequestObject{
			Params: messagingapi.GetFollowersParams{
				Limit: &limit,
			},
		})
		require.NoError(t, err)

		followers1, ok := resp1.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, followers1.UserIds, 2)
		assert.NotNil(t, followers1.Next)

		// Second page using the next token
		resp2, err := srv.GetFollowers(ctx, messagingapi.GetFollowersRequestObject{
			Params: messagingapi.GetFollowersParams{
				Start: followers1.Next,
				Limit: &limit,
			},
		})
		require.NoError(t, err)

		followers2, ok := resp2.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, followers2.UserIds, 2)
		assert.NotNil(t, followers2.Next)

		// Third page
		resp3, err := srv.GetFollowers(ctx, messagingapi.GetFollowersRequestObject{
			Params: messagingapi.GetFollowersParams{
				Start: followers2.Next,
				Limit: &limit,
			},
		})
		require.NoError(t, err)

		followers3, ok := resp3.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, followers3.UserIds, 1)
		assert.Nil(t, followers3.Next) // No more pages

		// Ensure no duplicate user IDs across pages
		allRetrievedIDs := append(append(followers1.UserIds, followers2.UserIds...), followers3.UserIds...)
		assert.Len(t, allRetrievedIDs, 5)
	})

	t.Run("success - empty followers list", func(t *testing.T) {
		// Create another bot with no followers
		emptyBot, err := dbClient.CreateBot(context.Background(), db.CreateBotParams{
			UserID:         "empty-bot-id",
			BasicID:        "empty-basic-id",
			ChatMode:       "bot",
			DisplayName:    "Empty Bot",
			MarkAsReadMode: "manual",
		})
		require.NoError(t, err)

		// Set the empty bot ID in context
		emptyCtx := auth.SetBotID(context.Background(), emptyBot.ID)

		resp, err := srv.GetFollowers(emptyCtx, messagingapi.GetFollowersRequestObject{})
		require.NoError(t, err)

		followers, ok := resp.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		assert.Len(t, followers.UserIds, 0)
		assert.Nil(t, followers.Next)
	})

	t.Run("success - limit capped at 1000", func(t *testing.T) {
		limit := int32(2000) // Over the max
		resp, err := srv.GetFollowers(ctx, messagingapi.GetFollowersRequestObject{
			Params: messagingapi.GetFollowersParams{
				Limit: &limit,
			},
		})
		require.NoError(t, err)

		followers, ok := resp.(messagingapi.GetFollowers200JSONResponse)
		assert.True(t, ok)
		// We only have 5 users, so we should get all of them
		assert.Len(t, followers.UserIds, 5)
		assert.Nil(t, followers.Next)
	})
}