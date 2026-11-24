package webhook_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/webhook"
)

type receivedRequest struct {
	signature string
	body      []byte
}

func newWebhookServer(t *testing.T, statusCode int) (*httptest.Server, func() []receivedRequest) {
	t.Helper()
	var mu sync.Mutex
	var requests []receivedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		requests = append(requests, receivedRequest{signature: r.Header.Get("X-Line-Signature"), body: body})
		mu.Unlock()
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(srv.Close)
	return srv, func() []receivedRequest {
		mu.Lock()
		defer mu.Unlock()
		return append([]receivedRequest(nil), requests...)
	}
}

func TestClient_SendEvents(t *testing.T) {
	t.Parallel()
	now := time.UnixMilli(1700000000000)

	t.Run("posts the events signed with the bot's channel secret", func(t *testing.T) {
		t.Parallel()
		srv, requests := newWebhookServer(t, http.StatusOK)
		event := webhook.NewTextMessageEvent("Uuser", "hello", now)

		err := webhook.NewClient().SendEvents(context.Background(), &webhook.SendEventsRequest{
			Endpoint:  srv.URL,
			BotUserID: "Ubot",
			Events:    []webhook.Event{event},
		})
		require.NoError(t, err)

		got := requests()
		require.Len(t, got, 1)
		mac := hmac.New(sha256.New, []byte("Ubot"))
		mac.Write(got[0].body)
		assert.Equal(t, base64.StdEncoding.EncodeToString(mac.Sum(nil)), got[0].signature)

		var payload struct {
			Destination string           `json:"destination"`
			Events      []map[string]any `json:"events"`
		}
		require.NoError(t, json.Unmarshal(got[0].body, &payload))
		assert.Equal(t, "Ubot", payload.Destination)
		require.Len(t, payload.Events, 1)
		assert.Equal(t, "message", payload.Events[0]["type"])
		assert.Equal(t, "active", payload.Events[0]["mode"])
		assert.Equal(t, float64(1700000000000), payload.Events[0]["timestamp"])
		assert.Equal(t, map[string]any{"type": "user", "userId": "Uuser"}, payload.Events[0]["source"])
		assert.Equal(t, event.ReplyToken(), payload.Events[0]["replyToken"])
		assert.Equal(t, event.ID(), payload.Events[0]["webhookEventId"])
	})

	t.Run("splits events into requests of at most 50", func(t *testing.T) {
		t.Parallel()
		srv, requests := newWebhookServer(t, http.StatusOK)
		events := make([]webhook.Event, 0, 51)
		for range 51 {
			events = append(events, webhook.NewFollowEvent("Uuser", now))
		}

		err := webhook.NewClient().SendEvents(context.Background(), &webhook.SendEventsRequest{Endpoint: srv.URL, BotUserID: "Ubot", Events: events})
		require.NoError(t, err)

		got := requests()
		require.Len(t, got, 2)
		counts := make([]int, 0, 2)
		for _, request := range got {
			var payload struct {
				Events []map[string]any `json:"events"`
			}
			require.NoError(t, json.Unmarshal(request.body, &payload))
			counts = append(counts, len(payload.Events))
		}
		assert.Equal(t, []int{50, 1}, counts)
	})

	t.Run("returns an error when the endpoint doesn't respond with a 2xx status", func(t *testing.T) {
		t.Parallel()
		srv, _ := newWebhookServer(t, http.StatusInternalServerError)

		err := webhook.NewClient().SendEvents(context.Background(), &webhook.SendEventsRequest{
			Endpoint:  srv.URL,
			BotUserID: "Ubot",
			Events:    []webhook.Event{webhook.NewFollowEvent("Uuser", now)},
		})

		assert.EqualError(t, err, "webhook request returned status code 500")
	})

	t.Run("returns an error when the endpoint is not set", func(t *testing.T) {
		t.Parallel()
		err := webhook.NewClient().SendEvents(context.Background(), &webhook.SendEventsRequest{
			BotUserID: "Ubot",
			Events:    []webhook.Event{webhook.NewFollowEvent("Uuser", now)},
		})

		assert.EqualError(t, err, "webhook endpoint is not set")
	})

	t.Run("an unfollow event has no reply token", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, webhook.NewUnfollowEvent("Uuser", now).ReplyToken())
	})
}
