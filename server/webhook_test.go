package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/api/adminapi"
	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
	"github.com/zero-color/line-messaging-api-emulator/server"
)

func TestGetWebhookEndpoint(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)
	ctx := context.Background()

	t.Run("returns webhook endpoint when configured", func(t *testing.T) {
		// Create a bot first
		createReq := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: "Test Bot",
			},
		}
		createResp, err := srv.CreateBot(ctx, createReq)
		require.NoError(t, err)
		createdBot, ok := createResp.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		// Get the bot from database to get its ID
		bot, err := dbClient.GetBotByUserID(ctx, createdBot.UserId)
		require.NoError(t, err)

		// Add bot ID to context
		botCtx := auth.SetBotID(ctx, bot.ID)

		// Set a webhook endpoint
		endpoint := "https://example.com/webhook"
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: &messagingapi.SetWebhookEndpointJSONRequestBody{
				Endpoint: endpoint,
			},
		}
		_, err = srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)

		// Get the webhook endpoint
		getReq := messagingapi.GetWebhookEndpointRequestObject{}
		resp, err := srv.GetWebhookEndpoint(botCtx, getReq)
		require.NoError(t, err)

		webhookResp, ok := resp.(messagingapi.GetWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, endpoint, webhookResp.Endpoint)
		assert.True(t, webhookResp.Active)
	})

	t.Run("returns empty endpoint when no webhook configured", func(t *testing.T) {
		// Create a bot without setting webhook
		createReq := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: "Test Bot 2",
			},
		}
		createResp, err := srv.CreateBot(ctx, createReq)
		require.NoError(t, err)
		createdBot, ok := createResp.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		// Get the bot from database to get its ID
		bot, err := dbClient.GetBotByUserID(ctx, createdBot.UserId)
		require.NoError(t, err)

		// Add bot ID to context
		botCtx := auth.SetBotID(ctx, bot.ID)

		// Get the webhook endpoint (should be empty)
		getReq := messagingapi.GetWebhookEndpointRequestObject{}
		resp, err := srv.GetWebhookEndpoint(botCtx, getReq)
		require.NoError(t, err)

		webhookResp, ok := resp.(messagingapi.GetWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "", webhookResp.Endpoint)
		assert.False(t, webhookResp.Active)
	})

	t.Run("returns error when bot does not exist", func(t *testing.T) {
		// Use a non-existent bot ID
		botCtx := auth.SetBotID(ctx, 99999)

		getReq := messagingapi.GetWebhookEndpointRequestObject{}
		resp, err := srv.GetWebhookEndpoint(botCtx, getReq)

		// When bot doesn't exist, we expect it to return empty endpoint
		require.NoError(t, err)
		webhookResp, ok := resp.(messagingapi.GetWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "", webhookResp.Endpoint)
		assert.False(t, webhookResp.Active)
	})
}

func TestSetWebhookEndpoint(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)
	ctx := context.Background()

	// Create a bot for all tests
	createReq := adminapi.CreateBotRequestObject{
		Body: &adminapi.CreateBotRequest{
			DisplayName: "Test Bot",
		},
	}
	createResp, err := srv.CreateBot(ctx, createReq)
	require.NoError(t, err)
	createdBot, ok := createResp.(adminapi.CreateBot201JSONResponse)
	require.True(t, ok)

	// Get the bot from database to get its ID
	bot, err := dbClient.GetBotByUserID(ctx, createdBot.UserId)
	require.NoError(t, err)
	botCtx := auth.SetBotID(ctx, bot.ID)

	t.Run("successfully sets webhook endpoint", func(t *testing.T) {
		endpoint := "https://example.com/webhook"
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: &messagingapi.SetWebhookEndpointJSONRequestBody{
				Endpoint: endpoint,
			},
		}

		resp, err := srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)
		_, ok := resp.(messagingapi.SetWebhookEndpoint200Response)
		require.True(t, ok)

		// Verify it was set
		getReq := messagingapi.GetWebhookEndpointRequestObject{}
		getResp, err := srv.GetWebhookEndpoint(botCtx, getReq)
		require.NoError(t, err)
		webhookResp, ok := getResp.(messagingapi.GetWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, endpoint, webhookResp.Endpoint)
	})

	t.Run("successfully updates webhook endpoint", func(t *testing.T) {
		// First set
		endpoint1 := "https://example.com/webhook1"
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: &messagingapi.SetWebhookEndpointJSONRequestBody{
				Endpoint: endpoint1,
			},
		}
		_, err := srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)

		// Update
		endpoint2 := "https://example.com/webhook2"
		setReq.Body.Endpoint = endpoint2
		resp, err := srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)
		_, ok := resp.(messagingapi.SetWebhookEndpoint200Response)
		require.True(t, ok)

		// Verify it was updated
		getReq := messagingapi.GetWebhookEndpointRequestObject{}
		getResp, err := srv.GetWebhookEndpoint(botCtx, getReq)
		require.NoError(t, err)
		webhookResp, ok := getResp.(messagingapi.GetWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, endpoint2, webhookResp.Endpoint)
	})

	t.Run("successfully clears webhook endpoint with empty string", func(t *testing.T) {
		// First set an endpoint
		endpoint := "https://example.com/webhook"
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: &messagingapi.SetWebhookEndpointJSONRequestBody{
				Endpoint: endpoint,
			},
		}
		_, err := srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)

		// Clear it
		setReq.Body.Endpoint = ""
		resp, err := srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)
		_, ok := resp.(messagingapi.SetWebhookEndpoint200Response)
		require.True(t, ok)

		// Verify it was cleared
		getReq := messagingapi.GetWebhookEndpointRequestObject{}
		getResp, err := srv.GetWebhookEndpoint(botCtx, getReq)
		require.NoError(t, err)
		webhookResp, ok := getResp.(messagingapi.GetWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.Equal(t, "", webhookResp.Endpoint)
		assert.True(t, webhookResp.Active) // Active remains true even with empty endpoint
	})

	t.Run("returns error for nil request body", func(t *testing.T) {
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: nil,
		}

		_, err := srv.SetWebhookEndpoint(botCtx, setReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request body is required")
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: &messagingapi.SetWebhookEndpointJSONRequestBody{
				Endpoint: "not-a-valid-url",
			},
		}

		_, err := srv.SetWebhookEndpoint(botCtx, setReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid webhook URL")
	})
}

func TestTestWebhookEndpoint(t *testing.T) {
	dbClient := db.NewTestDB(t)
	srv := server.New(dbClient)
	ctx := context.Background()

	// Create a bot for all tests
	createReq := adminapi.CreateBotRequestObject{
		Body: &adminapi.CreateBotRequest{
			DisplayName: "Test Bot",
		},
	}
	createResp, err := srv.CreateBot(ctx, createReq)
	require.NoError(t, err)
	createdBot, ok := createResp.(adminapi.CreateBot201JSONResponse)
	require.True(t, ok)

	// Get the bot from database to get its ID
	bot, err := dbClient.GetBotByUserID(ctx, createdBot.UserId)
	require.NoError(t, err)
	botCtx := auth.SetBotID(ctx, bot.ID)

	t.Run("successfully tests configured webhook endpoint", func(t *testing.T) {
		// Create a test server
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "test-signature", r.Header.Get("X-Line-Signature"))
			assert.Equal(t, "LineBotWebhook/2.0", r.Header.Get("User-Agent"))

			var payload map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&payload)
			require.NoError(t, err)
			assert.Equal(t, createdBot.UserId, payload["destination"])

			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		// Set the webhook endpoint
		setReq := messagingapi.SetWebhookEndpointRequestObject{
			Body: &messagingapi.SetWebhookEndpointJSONRequestBody{
				Endpoint: testServer.URL,
			},
		}
		_, err = srv.SetWebhookEndpoint(botCtx, setReq)
		require.NoError(t, err)

		// Test the webhook
		testReq := messagingapi.TestWebhookEndpointRequestObject{}
		resp, err := srv.TestWebhookEndpoint(botCtx, testReq)
		require.NoError(t, err)

		testResp, ok := resp.(messagingapi.TestWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.NotNil(t, testResp.Success)
		assert.True(t, *testResp.Success)
		assert.Equal(t, int32(200), testResp.StatusCode)
	})

	t.Run("tests provided endpoint URL", func(t *testing.T) {
		// Create a test server
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer testServer.Close()

		// Test with provided endpoint (not configured)
		testReq := messagingapi.TestWebhookEndpointRequestObject{
			Body: &messagingapi.TestWebhookEndpointJSONRequestBody{
				Endpoint: &testServer.URL,
			},
		}
		resp, err := srv.TestWebhookEndpoint(botCtx, testReq)
		require.NoError(t, err)

		testResp, ok := resp.(messagingapi.TestWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.NotNil(t, testResp.Success)
		assert.True(t, *testResp.Success)
		assert.Equal(t, int32(200), testResp.StatusCode)
	})

	t.Run("handles non-200 response codes", func(t *testing.T) {
		// Create a test server that returns 500
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer testServer.Close()

		testReq := messagingapi.TestWebhookEndpointRequestObject{
			Body: &messagingapi.TestWebhookEndpointJSONRequestBody{
				Endpoint: &testServer.URL,
			},
		}
		resp, err := srv.TestWebhookEndpoint(botCtx, testReq)
		require.NoError(t, err)

		testResp, ok := resp.(messagingapi.TestWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.NotNil(t, testResp.Success)
		assert.False(t, *testResp.Success)
		assert.Equal(t, int32(500), testResp.StatusCode)
	})

	t.Run("returns failure when no webhook configured", func(t *testing.T) {
		// Create a new bot without webhook
		createReq := adminapi.CreateBotRequestObject{
			Body: &adminapi.CreateBotRequest{
				DisplayName: "Test Bot No Webhook",
			},
		}
		createResp, err := srv.CreateBot(ctx, createReq)
		require.NoError(t, err)
		createdBot, ok := createResp.(adminapi.CreateBot201JSONResponse)
		require.True(t, ok)

		// Get the bot from database to get its ID
		bot, err := dbClient.GetBotByUserID(ctx, createdBot.UserId)
		require.NoError(t, err)
		newBotCtx := auth.SetBotID(ctx, bot.ID)

		testReq := messagingapi.TestWebhookEndpointRequestObject{}
		resp, err := srv.TestWebhookEndpoint(newBotCtx, testReq)
		require.NoError(t, err)

		testResp, ok := resp.(messagingapi.TestWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.NotNil(t, testResp.Success)
		assert.False(t, *testResp.Success)
		assert.Equal(t, int32(0), testResp.StatusCode)
		assert.Equal(t, "No webhook configured", testResp.Reason)
	})

	t.Run("handles invalid URL", func(t *testing.T) {
		invalidURL := "not-a-valid-url"
		testReq := messagingapi.TestWebhookEndpointRequestObject{
			Body: &messagingapi.TestWebhookEndpointJSONRequestBody{
				Endpoint: &invalidURL,
			},
		}

		resp, err := srv.TestWebhookEndpoint(botCtx, testReq)
		require.NoError(t, err)

		testResp, ok := resp.(messagingapi.TestWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.NotNil(t, testResp.Success)
		assert.False(t, *testResp.Success)
		assert.Equal(t, int32(0), testResp.StatusCode)
		assert.Equal(t, "Invalid URL", testResp.Reason)
	})

	t.Run("handles connection failure", func(t *testing.T) {
		// Use an unreachable URL
		unreachableURL := "http://localhost:99999/webhook"
		testReq := messagingapi.TestWebhookEndpointRequestObject{
			Body: &messagingapi.TestWebhookEndpointJSONRequestBody{
				Endpoint: &unreachableURL,
			},
		}

		resp, err := srv.TestWebhookEndpoint(botCtx, testReq)
		require.NoError(t, err)

		testResp, ok := resp.(messagingapi.TestWebhookEndpoint200JSONResponse)
		require.True(t, ok)
		assert.NotNil(t, testResp.Success)
		assert.False(t, *testResp.Success)
		assert.Equal(t, int32(0), testResp.StatusCode)
		assert.Equal(t, "Connection failed", testResp.Reason)
	})
}
