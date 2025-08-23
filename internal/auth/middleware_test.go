package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-color/line-messaging-api-emulator/db"
	"github.com/zero-color/line-messaging-api-emulator/internal/auth"
)

func TestMiddleware(t *testing.T) {
	// Setup test database
	dbClient := db.NewTestDB(t)
	
	// Create a test bot
	bot, err := dbClient.CreateBot(context.Background(), db.CreateBotParams{
		UserID:         "test-user-id",
		BasicID:        "@testbot",
		ChatMode:       "bot",
		DisplayName:    "Test Bot",
		MarkAsReadMode: "manual",
	})
	require.NoError(t, err)

	// Create the middleware - need to cast to *db.Queries
	queries, ok := dbClient.(*db.Queries)
	require.True(t, ok, "Failed to cast to *db.Queries")
	middleware := auth.Middleware(queries)

	// Create a test handler that will be wrapped by the middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify bot ID is in context
		botID := auth.GetBotID(r.Context())
		assert.Equal(t, bot.ID, botID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap the test handler with the middleware
	handler := middleware(testHandler)

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "missing authorization header")
	})

	t.Run("invalid authorization format - no bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid authorization format")
	})

	t.Run("invalid authorization format - not bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic dGVzdDp0ZXN0")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "invalid authorization format")
	})

	t.Run("bot not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer non-existent-user-id")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "bot not found")
	})

	t.Run("valid authorization with lowercase bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "bearer "+bot.UserID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("valid authorization with uppercase Bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+bot.UserID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("valid authorization with mixed case BEARER", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "BEARER "+bot.UserID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "success", rec.Body.String())
	})

	t.Run("authorization with extra spaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer  "+bot.UserID) // Extra space
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
		assert.Contains(t, rec.Body.String(), "bot not found")
	})
}

func TestGetBotID(t *testing.T) {
	t.Run("returns bot ID from context", func(t *testing.T) {
		ctx := context.Background()
		expectedID := int32(123)
		ctx = auth.SetBotID(ctx, expectedID)

		botID := auth.GetBotID(ctx)
		assert.Equal(t, expectedID, botID)
	})

	t.Run("panics when bot ID not in context", func(t *testing.T) {
		ctx := context.Background()
		
		assert.Panics(t, func() {
			auth.GetBotID(ctx)
		})
	})
}

func TestMiddlewareWithMultipleBots(t *testing.T) {
	// Setup test database
	dbClient := db.NewTestDB(t)
	
	// Create multiple test bots
	bot1, err := dbClient.CreateBot(context.Background(), db.CreateBotParams{
		UserID:         "bot1-user-id",
		BasicID:        "@bot1",
		ChatMode:       "bot",
		DisplayName:    "Bot 1",
		MarkAsReadMode: "manual",
	})
	require.NoError(t, err)

	bot2, err := dbClient.CreateBot(context.Background(), db.CreateBotParams{
		UserID:         "bot2-user-id",
		BasicID:        "@bot2",
		ChatMode:       "chat",
		DisplayName:    "Bot 2",
		MarkAsReadMode: "auto",
	})
	require.NoError(t, err)

	// Create the middleware - need to cast to *db.Queries
	queries, ok := dbClient.(*db.Queries)
	require.True(t, ok, "Failed to cast to *db.Queries")
	middleware := auth.Middleware(queries)

	// Create a test handler that captures the bot ID
	var capturedBotID int32
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBotID = auth.GetBotID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware(testHandler)

	t.Run("authenticates with first bot", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+bot1.UserID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, bot1.ID, capturedBotID)
	})

	t.Run("authenticates with second bot", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+bot2.UserID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, bot2.ID, capturedBotID)
	})
}