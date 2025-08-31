package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

// GetWebhookEndpoint gets the webhook endpoint URL
func (s *server) GetWebhookEndpoint(ctx context.Context, _ messagingapi.GetWebhookEndpointRequestObject) (messagingapi.GetWebhookEndpointResponseObject, error) {
	botID := auth.GetBotID(ctx)

	webhook, err := s.db.GetWebhook(ctx, botID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No webhook configured, return empty endpoint with active=false
			return messagingapi.GetWebhookEndpoint200JSONResponse{
				Endpoint: "",
				Active:   false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	return messagingapi.GetWebhookEndpoint200JSONResponse{
		Endpoint: webhook.Endpoint,
		Active:   webhook.Active,
	}, nil
}

// SetWebhookEndpoint sets the webhook endpoint URL
func (s *server) SetWebhookEndpoint(ctx context.Context, request messagingapi.SetWebhookEndpointRequestObject) (messagingapi.SetWebhookEndpointResponseObject, error) {
	if request.Body == nil {
		return nil, fmt.Errorf("request body is required")
	}

	botID := auth.GetBotID(ctx)

	// Validate the webhook URL
	if request.Body.Endpoint != "" {
		if _, err := url.ParseRequestURI(request.Body.Endpoint); err != nil {
			return nil, fmt.Errorf("invalid webhook URL: %w", err)
		}
	}

	err := s.db.UpsertWebhook(ctx, db.UpsertWebhookParams{
		BotID:    botID,
		Endpoint: request.Body.Endpoint,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to set webhook: %w", err)
	}

	return messagingapi.SetWebhookEndpoint200Response{}, nil
}

// TestWebhookEndpoint tests the webhook endpoint
func (s *server) TestWebhookEndpoint(ctx context.Context, request messagingapi.TestWebhookEndpointRequestObject) (messagingapi.TestWebhookEndpointResponseObject, error) {
	botID := auth.GetBotID(ctx)

	// Get the bot to get its user ID for the test payload
	bot, err := s.db.GetBot(ctx, botID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot: %w", err)
	}

	var endpoint string
	if request.Body != nil && request.Body.Endpoint != nil {
		// Test the provided endpoint
		endpoint = *request.Body.Endpoint
	} else {
		// Test the currently configured endpoint
		webhook, err := s.db.GetWebhook(ctx, botID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				falseVal := false
				timestamp := time.Now()
				return messagingapi.TestWebhookEndpoint200JSONResponse{
					Success:    &falseVal,
					Timestamp:  timestamp,
					StatusCode: 0,
					Reason:     "No webhook configured",
					Detail:     "No webhook URL is configured for this bot",
				}, nil
			}
			return nil, fmt.Errorf("failed to get webhook: %w", err)
		}

		endpoint = webhook.Endpoint
	}

	// Validate the URL
	parsedURL, err := url.ParseRequestURI(endpoint)
	if err != nil {
		falseVal := false
		timestamp := time.Now()
		return messagingapi.TestWebhookEndpoint200JSONResponse{
			Success:    &falseVal,
			Timestamp:  timestamp,
			StatusCode: 0,
			Reason:     "Invalid URL",
			Detail:     fmt.Sprintf("Invalid webhook URL: %v", err),
		}, nil
	}

	// Create a test event payload
	testPayload := map[string]interface{}{
		"destination": bot.UserID,
		"events":      []interface{}{},
	}

	payloadBytes, _ := json.Marshal(testPayload)

	// Make a POST request to the webhook URL
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", parsedURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		falseVal := false
		timestamp := time.Now()
		return messagingapi.TestWebhookEndpoint200JSONResponse{
			Success:    &falseVal,
			Timestamp:  timestamp,
			StatusCode: 0,
			Reason:     "Request creation failed",
			Detail:     fmt.Sprintf("Failed to create request: %v", err),
		}, nil
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Line-Signature", "test-signature")
	req.Header.Set("User-Agent", "LineBotWebhook/2.0")

	resp, err := client.Do(req)
	if err != nil {
		falseVal := false
		timestamp := time.Now()
		return messagingapi.TestWebhookEndpoint200JSONResponse{
			Success:    &falseVal,
			Timestamp:  timestamp,
			StatusCode: 0,
			Reason:     "Connection failed",
			Detail:     fmt.Sprintf("Failed to connect to webhook: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	timestamp := time.Now()

	return messagingapi.TestWebhookEndpoint200JSONResponse{
		Success:    &success,
		Timestamp:  timestamp,
		StatusCode: int32(resp.StatusCode),
		Reason:     resp.Status,
		Detail:     fmt.Sprintf("%d", resp.StatusCode),
	}, nil
}
