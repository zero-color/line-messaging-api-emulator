package server_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageQuota(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	text := func(texts ...string) []map[string]any {
		messages := make([]map[string]any, 0, len(texts))
		for _, text := range texts {
			messages = append(messages, map[string]any{"type": "text", "text": text})
		}
		return messages
	}
	consumption := func(t *testing.T, channelAccessToken string) int64 {
		t.Helper()
		var resp struct {
			TotalUsage int64 `json:"totalUsage"`
		}
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodGet, "/v2/bot/message/quota/consumption", channelAccessToken, nil, &resp))
		return resp.TotalUsage
	}

	t.Run("a bot has no message limit", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)

		var quota map[string]any
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodGet, "/v2/bot/message/quota", bot.UserID, nil, &quota))

		assert.Equal(t, map[string]any{"type": "none"}, quota)
	})

	t.Run("consumption is zero before anything is sent", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)

		assert.Equal(t, int64(0), consumption(t, bot.UserID))
	})

	t.Run("consumption counts each message object per recipient and leaves out replies and other bots", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, true)
		otherBot := createTestBot(ctx, t, dbClient, false)
		first := createTestUser(ctx, t, dbClient, &bot)
		second := createTestUser(ctx, t, dbClient, &bot)
		send := func(channelAccessToken string, path string, body map[string]any) {
			t.Helper()
			require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodPost, path, channelAccessToken, body, nil))
		}

		// push: 2 messages to 1 user = 2
		send(bot.UserID, "/v2/bot/message/push", map[string]any{"to": first.UserID, "messages": text("a", "b")})
		// multicast: 1 message to 2 users = 2
		send(bot.UserID, "/v2/bot/message/multicast", map[string]any{"to": []string{first.UserID, second.UserID}, "messages": text("c")})
		// broadcast: 1 message to 2 followers = 2
		send(bot.UserID, "/v2/bot/message/broadcast", map[string]any{"messages": text("d")})
		// reply: free
		var event struct {
			ReplyToken string `json:"replyToken"`
		}
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodPost, "/admin/bots/"+bot.UserID+"/users/"+first.UserID+"/events", "",
			map[string]any{"type": "message", "message": map[string]any{"text": "hello"}}, &event))
		send(bot.UserID, "/v2/bot/message/reply", map[string]any{"replyToken": event.ReplyToken, "messages": text("e")})
		send(otherBot.UserID, "/v2/bot/message/push", map[string]any{"to": first.UserID, "messages": text("f")})

		assert.Equal(t, int64(6), consumption(t, bot.UserID))
		assert.Equal(t, int64(1), consumption(t, otherBot.UserID))
	})
}

func TestInsight(t *testing.T) {
	ctx, dbClient, baseURL := newRichMenuTestServer(t)

	t.Run("the number of followers is the bot's current follower count", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)
		otherBot := createTestBot(ctx, t, dbClient, false)
		createTestUser(ctx, t, dbClient, &bot)
		createTestUser(ctx, t, dbClient, &bot)
		createTestUser(ctx, t, dbClient, &otherBot)

		var got map[string]any
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodGet, "/v2/bot/insight/followers?date=20260101", bot.UserID, nil, &got))

		assert.Equal(t, map[string]any{"status": "ready", "followers": float64(2), "targetedReaches": float64(2), "blocks": float64(0)}, got)
	})

	t.Run("a bot without followers reports zero", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)

		var got map[string]any
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodGet, "/v2/bot/insight/followers", bot.UserID, nil, &got))

		assert.Equal(t, map[string]any{"status": "ready", "followers": float64(0), "targetedReaches": float64(0), "blocks": float64(0)}, got)
	})

	t.Run("demographics are reported as unavailable", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)

		var got map[string]any
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodGet, "/v2/bot/insight/demographic", bot.UserID, nil, &got))

		assert.Equal(t, map[string]any{"available": false}, got)
	})

	t.Run("statistics per aggregation unit are empty with null overview values", func(t *testing.T) {
		bot := createTestBot(ctx, t, dbClient, false)

		var got map[string]any
		path := "/v2/bot/insight/message/event/aggregation?customAggregationUnit=promotion_a&from=20260101&to=20260131"
		require.Equal(t, http.StatusOK, callAPI(t, baseURL, http.MethodGet, path, bot.UserID, nil, &got))

		assert.Equal(t, map[string]any{
			"overview": map[string]any{"uniqueImpression": nil, "uniqueClick": nil, "uniqueMediaPlayed": nil, "uniqueMediaPlayed100Percent": nil},
			"messages": []any{},
			"clicks":   []any{},
		}, got)
	})

	t.Run("the insight API requires a channel access token", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, callAPI(t, baseURL, http.MethodGet, "/v2/bot/insight/followers", "", nil, nil))
	})
}
