package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
	"github.com/zero-color/line-messaging-api-emulator/server"
	"github.com/zero-color/line-messaging-api-emulator/webhook"
)

const testWebhookEndpoint = "https://example.com/webhook"

func createTestBot(ctx context.Context, t *testing.T, dbClient db.DB, withWebhook bool) db.Bot {
	t.Helper()
	id := uuid.NewString()
	bot, err := dbClient.Queries().CreateBot(ctx, db.CreateBotParams{
		UserID:         "U" + id,
		BasicID:        "@" + id,
		ChatMode:       "bot",
		DisplayName:    "Test Bot",
		MarkAsReadMode: "manual",
	})
	require.NoError(t, err)
	if withWebhook {
		require.NoError(t, dbClient.Queries().UpsertWebhook(ctx, db.UpsertWebhookParams{BotID: bot.ID, Endpoint: testWebhookEndpoint}))
	}
	return bot
}

func createTestUser(ctx context.Context, t *testing.T, dbClient db.DB, followedBot *db.Bot) db.User {
	t.Helper()
	user, err := dbClient.Queries().CreateUser(ctx, db.CreateUserParams{
		UserID:      "U" + uuid.NewString(),
		DisplayName: "Test User",
	})
	require.NoError(t, err)
	if followedBot != nil {
		_, err := dbClient.Queries().CreateBotFollower(ctx, db.CreateBotFollowerParams{BotID: followedBot.ID, UserID: user.ID})
		require.NoError(t, err)
	}
	return user
}

func textMessages() []messagingapi.Message {
	return []messagingapi.Message{{Type: "text"}}
}

// callAPI sends a JSON request to the router and decodes the JSON response into out when it isn't nil.
// channelAccessToken is omitted for the admin API.
func callAPI(t *testing.T, baseURL string, method string, path string, channelAccessToken string, body any, out any) int {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, baseURL+path, reqBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if channelAccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+channelAccessToken)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if out != nil {
		require.NoError(t, json.Unmarshal(respBody, out), string(respBody))
	}
	return resp.StatusCode
}

func TestSendUserEvent(t *testing.T) {
	dbClient := db.NewTestDB(t)
	ctx := db.ToContext(context.Background(), dbClient)

	sendEvent := func(t *testing.T, srv server.Server, bot db.Bot, user db.User, body adminapi.SendUserEventRequest) adminapi.SendUserEventResponseObject {
		t.Helper()
		resp, err := srv.SendUserEvent(ctx, adminapi.SendUserEventRequestObject{BotId: bot.UserID, UserId: user.UserID, Body: &body})
		require.NoError(t, err)
		return resp
	}

	t.Run("follow adds the follower and delivers a follow event", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		resp := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Follow})

		success, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)
		assert.True(t, success.Delivered)
		assert.Nil(t, success.Detail)

		isFollower, err := dbClient.Queries().IsBotFollower(ctx, db.IsBotFollowerParams{BotID: bot.ID, UserID: user.UserID})
		require.NoError(t, err)
		assert.True(t, isFollower)

		calls := webhookClient.SendEventsCalls()
		require.Len(t, calls, 1)
		assert.Equal(t, testWebhookEndpoint, calls[0].Endpoint)
		assert.Equal(t, bot.UserID, calls[0].BotUserID)
		require.Len(t, calls[0].Events, 1)
		event := calls[0].Events[0]
		assert.Equal(t, "follow", event["type"])
		assert.Equal(t, map[string]any{"type": "user", "userId": user.UserID}, event["source"])
		assert.Equal(t, success.WebhookEventId, event.ID())
		assert.Equal(t, lo.FromPtr(success.ReplyToken), event.ReplyToken())
		assert.NotEmpty(t, event.ReplyToken())
	})

	t.Run("follow is rejected when the user already follows the bot", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)

		resp := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Follow})

		failure, ok := resp.(adminapi.SendUserEvent400JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "The user already follows the bot", failure.Error.Message)
		assert.Empty(t, webhookClient.SendEventsCalls())
	})

	t.Run("unfollow removes the follower and delivers an event without a reply token", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)

		resp := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Unfollow})

		success, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)
		assert.True(t, success.Delivered)
		assert.Nil(t, success.ReplyToken)

		isFollower, err := dbClient.Queries().IsBotFollower(ctx, db.IsBotFollowerParams{BotID: bot.ID, UserID: user.UserID})
		require.NoError(t, err)
		assert.False(t, isFollower)

		calls := webhookClient.SendEventsCalls()
		require.Len(t, calls, 1)
		require.Len(t, calls[0].Events, 1)
		assert.Equal(t, "unfollow", calls[0].Events[0]["type"])
	})

	t.Run("unfollow is rejected when the user doesn't follow the bot", func(t *testing.T) {
		srv := server.New(webhook.NewMockClient())
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		resp := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Unfollow})

		failure, ok := resp.(adminapi.SendUserEvent400JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "The user doesn't follow the bot", failure.Error.Message)
	})

	t.Run("message delivers a text message event", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)

		body := adminapi.SendUserEventRequest{Type: adminapi.Message}
		body.Message = &struct {
			Text string `json:"text"`
		}{Text: "hello"}
		resp := sendEvent(t, srv, bot, user, body)

		success, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)
		assert.True(t, success.Delivered)

		calls := webhookClient.SendEventsCalls()
		require.Len(t, calls, 1)
		require.Len(t, calls[0].Events, 1)
		event := calls[0].Events[0]
		assert.Equal(t, "message", event["type"])
		message, ok := event["message"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "text", message["type"])
		assert.Equal(t, "hello", message["text"])
		assert.NotEmpty(t, message["id"])
	})

	t.Run("postback delivers the data and params", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)

		body := adminapi.SendUserEventRequest{Type: adminapi.Postback}
		body.Postback = &struct {
			Data   string             `json:"data"`
			Params *map[string]string `json:"params,omitempty"`
		}{Data: "action=buy", Params: &map[string]string{"date": "2026-01-01"}}
		resp := sendEvent(t, srv, bot, user, body)

		_, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)

		calls := webhookClient.SendEventsCalls()
		require.Len(t, calls, 1)
		require.Len(t, calls[0].Events, 1)
		event := calls[0].Events[0]
		assert.Equal(t, "postback", event["type"])
		assert.Equal(t, map[string]any{"data": "action=buy", "params": map[string]string{"date": "2026-01-01"}}, event["postback"])
	})

	t.Run("message and postback are rejected when the user doesn't follow the bot", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		messageBody := adminapi.SendUserEventRequest{Type: adminapi.Message}
		messageBody.Message = &struct {
			Text string `json:"text"`
		}{Text: "hello"}
		postbackBody := adminapi.SendUserEventRequest{Type: adminapi.Postback}
		postbackBody.Postback = &struct {
			Data   string             `json:"data"`
			Params *map[string]string `json:"params,omitempty"`
		}{Data: "action=buy"}

		for _, body := range []adminapi.SendUserEventRequest{messageBody, postbackBody} {
			failure, ok := sendEvent(t, srv, bot, user, body).(adminapi.SendUserEvent400JSONResponse)
			require.True(t, ok)
			assert.Equal(t, "The user doesn't follow the bot", failure.Error.Message)
		}
		assert.Empty(t, webhookClient.SendEventsCalls())
	})

	t.Run("message without text and postback without data are rejected", func(t *testing.T) {
		srv := server.New(webhook.NewMockClient())
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)

		messageFailure, ok := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Message}).(adminapi.SendUserEvent400JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "message.text is required for a message event", messageFailure.Error.Message)

		postbackFailure, ok := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Postback}).(adminapi.SendUserEvent400JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "postback.data is required for a postback event", postbackFailure.Error.Message)
	})

	t.Run("unknown event type is rejected", func(t *testing.T) {
		srv := server.New(webhook.NewMockClient())
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)

		failure, ok := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: "join"}).(adminapi.SendUserEvent400JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "Unknown event type: join", failure.Error.Message)
	})

	t.Run("the action is applied but reported undelivered when no webhook endpoint is set", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, false)
		user := createTestUser(ctx, t, dbClient, nil)

		resp := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Follow})

		success, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)
		assert.False(t, success.Delivered)
		assert.Equal(t, "webhook endpoint is not set", lo.FromPtr(success.Detail))
		assert.Empty(t, webhookClient.SendEventsCalls())

		isFollower, err := dbClient.Queries().IsBotFollower(ctx, db.IsBotFollowerParams{BotID: bot.ID, UserID: user.UserID})
		require.NoError(t, err)
		assert.True(t, isFollower)
	})

	t.Run("a webhook failure is reported in the response", func(t *testing.T) {
		webhookClient := webhook.NewMockClient()
		webhookClient.SendEventsFunc = func(context.Context, *webhook.SendEventsRequest) error {
			return errors.New("webhook request returned status code 500")
		}
		srv := server.New(webhookClient)
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		resp := sendEvent(t, srv, bot, user, adminapi.SendUserEventRequest{Type: adminapi.Follow})

		success, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)
		assert.False(t, success.Delivered)
		assert.Equal(t, "webhook request returned status code 500", lo.FromPtr(success.Detail))
	})

	t.Run("unknown bot or user returns not found", func(t *testing.T) {
		srv := server.New(webhook.NewMockClient())
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		_, ok := sendEvent(t, srv, db.Bot{UserID: "Uunknown"}, user, adminapi.SendUserEventRequest{Type: adminapi.Follow}).(adminapi.SendUserEvent404JSONResponse)
		assert.True(t, ok)
		_, ok = sendEvent(t, srv, bot, db.User{UserID: "Uunknown"}, adminapi.SendUserEventRequest{Type: adminapi.Follow}).(adminapi.SendUserEvent404JSONResponse)
		assert.True(t, ok)
	})
}

func TestListUserMessages(t *testing.T) {
	dbClient := db.NewTestDB(t)
	ctx := db.ToContext(context.Background(), dbClient)
	apiServer := httptest.NewServer(server.NewRouter(dbClient, server.New(webhook.NewMockClient())))
	t.Cleanup(apiServer.Close)

	type sendRequest struct {
		To         any              `json:"to,omitempty"`
		ReplyToken string           `json:"replyToken,omitempty"`
		Messages   []map[string]any `json:"messages"`
	}
	text := func(text string) []map[string]any {
		return []map[string]any{{"type": "text", "text": text}}
	}
	send := func(t *testing.T, bot db.Bot, path string, body sendRequest) {
		t.Helper()
		status := callAPI(t, apiServer.URL, http.MethodPost, path, bot.UserID, body, nil)
		require.Equal(t, http.StatusOK, status)
	}
	type received struct {
		SendType string           `json:"sendType"`
		Messages []map[string]any `json:"messages"`
	}
	listMessages := func(t *testing.T, bot db.Bot, user db.User) []received {
		t.Helper()
		var resp struct {
			Messages []received `json:"messages"`
		}
		status := callAPI(t, apiServer.URL, http.MethodGet, "/admin/bots/"+bot.UserID+"/users/"+user.UserID+"/messages", "", nil, &resp)
		require.Equal(t, http.StatusOK, status)
		return resp.Messages
	}

	t.Run("returns only what the user received, oldest first, with the full message objects", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		otherBot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)
		otherUser := createTestUser(ctx, t, dbClient, &bot)

		send(t, bot, "/v2/bot/message/push", sendRequest{To: user.UserID, Messages: text("push")})
		send(t, bot, "/v2/bot/message/push", sendRequest{To: otherUser.UserID, Messages: text("push to other user")})
		send(t, bot, "/v2/bot/message/multicast", sendRequest{To: []string{otherUser.UserID, user.UserID}, Messages: text("multicast")})
		send(t, bot, "/v2/bot/message/multicast", sendRequest{To: []string{otherUser.UserID}, Messages: text("multicast to other user")})
		send(t, bot, "/v2/bot/message/broadcast", sendRequest{Messages: text("broadcast")})
		send(t, otherBot, "/v2/bot/message/push", sendRequest{To: user.UserID, Messages: text("push from other bot")})

		assert.Equal(t, []received{
			{SendType: "push", Messages: text("push")},
			{SendType: "multicast", Messages: text("multicast")},
			{SendType: "broadcast", Messages: text("broadcast")},
		}, listMessages(t, bot, user))
	})

	t.Run("a reply is attributed to the user whose event issued the reply token", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)
		otherUser := createTestUser(ctx, t, dbClient, &bot)

		var event struct {
			ReplyToken string `json:"replyToken"`
		}
		status := callAPI(t, apiServer.URL, http.MethodPost, "/admin/bots/"+bot.UserID+"/users/"+user.UserID+"/events", "",
			map[string]any{"type": "message", "message": map[string]any{"text": "hello"}}, &event)
		require.Equal(t, http.StatusOK, status)
		require.NotEmpty(t, event.ReplyToken)

		send(t, bot, "/v2/bot/message/reply", sendRequest{ReplyToken: event.ReplyToken, Messages: text("reply")})

		assert.Equal(t, []received{{SendType: "reply", Messages: text("reply")}}, listMessages(t, bot, user))
		assert.Empty(t, listMessages(t, bot, otherUser))
	})

	t.Run("a send with an empty retry key header is accepted", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)
		body, err := json.Marshal(sendRequest{To: user.UserID, Messages: text("push")})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, apiServer.URL+"/v2/bot/message/push", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+bot.UserID)
		req.Header.Set("X-Line-Retry-Key", "")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, []received{{SendType: "push", Messages: text("push")}}, listMessages(t, bot, user))
	})

	t.Run("a broadcast is not listed for a user who doesn't follow the bot", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		send(t, bot, "/v2/bot/message/broadcast", sendRequest{Messages: text("broadcast")})

		assert.Empty(t, listMessages(t, bot, user))
	})

	t.Run("each entry has an ID and the time it was sent", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, &bot)
		send(t, bot, "/v2/bot/message/push", sendRequest{To: user.UserID, Messages: text("push")})

		var resp adminapi.ListUserMessagesResponse
		status := callAPI(t, apiServer.URL, http.MethodGet, "/admin/bots/"+bot.UserID+"/users/"+user.UserID+"/messages", "", nil, &resp)
		require.Equal(t, http.StatusOK, status)
		require.Len(t, resp.Messages, 1)
		assert.NotEmpty(t, resp.Messages[0].Id)
		assert.False(t, resp.Messages[0].SentAt.IsZero())
	})

	t.Run("unknown bot or user returns not found", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		user := createTestUser(ctx, t, dbClient, nil)

		assert.Equal(t, http.StatusNotFound, callAPI(t, apiServer.URL, http.MethodGet, "/admin/bots/Uunknown/users/"+user.UserID+"/messages", "", nil, nil))
		assert.Equal(t, http.StatusNotFound, callAPI(t, apiServer.URL, http.MethodGet, "/admin/bots/"+bot.UserID+"/users/Uunknown/messages", "", nil, nil))
	})
}

func TestReplyMessage(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(webhook.NewMockClient())
	ctx := db.ToContext(context.Background(), dbClient)

	issueReplyToken := func(t *testing.T, bot db.Bot) string {
		t.Helper()
		user := createTestUser(ctx, t, dbClient, nil)
		resp, err := srv.SendUserEvent(ctx, adminapi.SendUserEventRequestObject{BotId: bot.UserID, UserId: user.UserID, Body: &adminapi.SendUserEventRequest{Type: adminapi.Follow}})
		require.NoError(t, err)
		success, ok := resp.(adminapi.SendUserEvent200JSONResponse)
		require.True(t, ok)
		require.NotNil(t, success.ReplyToken)
		return *success.ReplyToken
	}

	t.Run("a reply token can be used once", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		replyToken := issueReplyToken(t, bot)
		req := messagingapi.ReplyMessageRequestObject{Body: &messagingapi.ReplyMessageRequest{ReplyToken: replyToken, Messages: textMessages()}}

		resp, err := srv.ReplyMessage(auth.SetBot(ctx, &bot), req)
		require.NoError(t, err)
		success, ok := resp.(messagingapi.ReplyMessage200JSONResponse)
		require.True(t, ok)
		assert.Len(t, success.SentMessages, 1)

		_, err = srv.ReplyMessage(auth.SetBot(ctx, &bot), req)
		var validationErr *server.ValidationError
		require.ErrorAs(t, err, &validationErr)
	})

	t.Run("a reply token that was never issued is rejected", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)

		_, err := srv.ReplyMessage(auth.SetBot(ctx, &bot), messagingapi.ReplyMessageRequestObject{Body: &messagingapi.ReplyMessageRequest{ReplyToken: "unknown", Messages: textMessages()}})

		var validationErr *server.ValidationError
		require.ErrorAs(t, err, &validationErr)
	})

	t.Run("a reply token issued for another bot is rejected", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		otherBot := createTestBot(ctx, t, dbClient, true)
		replyToken := issueReplyToken(t, otherBot)

		_, err := srv.ReplyMessage(auth.SetBot(ctx, &bot), messagingapi.ReplyMessageRequestObject{Body: &messagingapi.ReplyMessageRequest{ReplyToken: replyToken, Messages: textMessages()}})

		var validationErr *server.ValidationError
		require.ErrorAs(t, err, &validationErr)
	})
}
