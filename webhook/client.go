package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
)

// FollowEventRequest contains the data needed to send follow events
type FollowEventRequest struct {
	Endpoint  string
	BotUserID string
	Users     []FollowUser
}

// FollowUser represents a user who followed the bot
type FollowUser struct {
	UserID string
}

// Client is the interface for webhook operations
type Client interface {
	// SendFollowEvents sends follow events to the webhook endpoint
	SendFollowEvents(ctx context.Context, req *FollowEventRequest) error
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

// SendFollowEvents sends follow events to the configured webhook endpoint
func (c *client) SendFollowEvents(ctx context.Context, req *FollowEventRequest) error {
	if req.Endpoint == "" {
		return nil
	}

	events := make([]map[string]any, 0, len(req.Users))
	now := time.Now()

	for _, user := range req.Users {
		events = append(events, buildFollowEvent(user.UserID, now))
	}

	var errs []error
	for _, eventChunk := range lo.Chunk(events, 50) {
		payload := map[string]any{
			"destination": req.BotUserID,
			"events":      eventChunk,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to marshal webhook payload: %w", err))
			continue
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, req.Endpoint, bytes.NewBuffer(payloadBytes))
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create webhook request: %w", err))
			continue
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-Line-Signature", computeSignature(req.BotUserID, payloadBytes))
		httpReq.Header.Set("User-Agent", "LineBotWebhook/2.0")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to send webhook request: %w", err))
			continue
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				errs = append(errs, fmt.Errorf("webhook request returned status code %d", resp.StatusCode))
			}
		}()
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors occurred while sending follow events: %v", errs)
	}
	return nil
}

// buildFollowEvent creates a follow event payload for a user
func buildFollowEvent(userID string, timestamp time.Time) map[string]any {
	return map[string]any{
		"replyToken": strings.ReplaceAll(uuid.New().String(), "-", ""),
		"type":       "follow",
		"mode":       "active",
		"timestamp":  timestamp.UnixMilli(),
		"source": map[string]any{
			"type":   "user",
			"userId": userID,
		},
		"webhookEventId": ulid.Make().String(),
		"deliveryContext": map[string]any{
			"isRedelivery": false,
		},
		"follow": map[string]any{
			"isUnblocked": false,
		},
	}
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
