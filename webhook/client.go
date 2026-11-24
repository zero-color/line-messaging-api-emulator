package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
)

// Event is a webhook event object as defined by the Messaging API.
// See: https://developers.line.biz/en/reference/messaging-api/#webhook-event-objects
type Event map[string]any

// ID returns the webhook event ID.
func (e Event) ID() string {
	id, _ := e["webhookEventId"].(string)
	return id
}

// ReplyToken returns the reply token, or an empty string for events that can't be replied to.
func (e Event) ReplyToken() string {
	token, _ := e["replyToken"].(string)
	return token
}

// SendEventsRequest contains the data needed to deliver webhook events
type SendEventsRequest struct {
	Endpoint  string
	BotUserID string
	Events    []Event
}

// Client is the interface for webhook operations
type Client interface {
	// SendEvents delivers the events to the webhook endpoint.
	// It returns an error when the endpoint can't be reached or doesn't respond with a 2xx status.
	SendEvents(ctx context.Context, req *SendEventsRequest) error
}

type client struct {
	httpClient *http.Client
}

// NewClient creates a new webhook client
func NewClient() Client {
	return &client{
		httpClient: &http.Client{
			// Same as Line Messaging API webhook client timeout
			// https://developers.line.biz/ja/docs/messaging-api/check-webhook-error-statistics/#reason-request-timeout
			Timeout: 2 * time.Second,
		},
	}
}

// maxEventsPerRequest caps how many events are batched into a single webhook request.
const maxEventsPerRequest = 50

// SendEvents sends the events to the webhook endpoint in batches
func (c *client) SendEvents(ctx context.Context, req *SendEventsRequest) error {
	if req.Endpoint == "" {
		return errors.New("webhook endpoint is not set")
	}

	var errs []error
	for _, eventChunk := range lo.Chunk(req.Events, maxEventsPerRequest) {
		if err := c.post(ctx, req.Endpoint, req.BotUserID, eventChunk); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (c *client) post(ctx context.Context, endpoint string, botUserID string, events []Event) error {
	payloadBytes, err := json.Marshal(map[string]any{
		"destination": botUserID,
		"events":      events,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Line-Signature", computeSignature(botUserID, payloadBytes))
	httpReq.Header.Set("User-Agent", "LineBotWebhook/2.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook request returned status code %d", resp.StatusCode)
	}
	return nil
}

// NewFollowEvent creates the event sent when a user adds the bot as a friend
func NewFollowEvent(userID string, timestamp time.Time) Event {
	event := newUserEvent("follow", userID, timestamp)
	event["replyToken"] = newReplyToken()
	event["follow"] = map[string]any{"isUnblocked": false}
	return event
}

// NewUnfollowEvent creates the event sent when a user blocks the bot.
// Unfollow events have no reply token.
func NewUnfollowEvent(userID string, timestamp time.Time) Event {
	return newUserEvent("unfollow", userID, timestamp)
}

// NewTextMessageEvent creates the event sent when a user sends a text message to the bot
func NewTextMessageEvent(userID string, text string, timestamp time.Time) Event {
	messageID := strconv.FormatInt(timestamp.UnixNano(), 10)
	event := newUserEvent("message", userID, timestamp)
	event["replyToken"] = newReplyToken()
	event["message"] = map[string]any{
		"type":       "text",
		"id":         messageID,
		"quoteToken": newReplyToken(),
		"text":       text,
	}
	return event
}

// NewPostbackEvent creates the event sent when a user triggers a postback action.
// params is included only when it isn't empty, as for datetime picker and rich menu switch actions.
func NewPostbackEvent(userID string, data string, params map[string]string, timestamp time.Time) Event {
	postback := map[string]any{"data": data}
	if len(params) > 0 {
		postback["params"] = params
	}
	event := newUserEvent("postback", userID, timestamp)
	event["replyToken"] = newReplyToken()
	event["postback"] = postback
	return event
}

func newUserEvent(eventType string, userID string, timestamp time.Time) Event {
	return Event{
		"type":      eventType,
		"mode":      "active",
		"timestamp": timestamp.UnixMilli(),
		"source": map[string]any{
			"type":   "user",
			"userId": userID,
		},
		"webhookEventId": ulid.Make().String(),
		"deliveryContext": map[string]any{
			"isRedelivery": false,
		},
	}
}

func newReplyToken() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// computeSignature computes the X-Line-Signature header value.
// The signature is Base64(HMAC-SHA256(channelSecret, body)). This emulator uses
// the bot's user ID as the channel secret, mirroring the bearer-token convention
// in internal/auth.
// See: https://developers.line.biz/ja/docs/messaging-api/verify-webhook-signature/
func computeSignature(channelSecret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(channelSecret))
	mac.Write(body)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
