package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/zero-color/line-messaging-api-emulator/api/messagingapi"
)

type rawBodyContextKey struct{}

// RawBodyMiddleware keeps the request body in the context so that handlers can read fields
// the generated request types drop.
func RawBodyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.Body == http.NoBody {
				next.ServeHTTP(w, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), rawBodyContextKey{}, body)))
		})
	}
}

// rawRequestBody returns the request body kept by RawBodyMiddleware
func rawRequestBody(ctx context.Context) ([]byte, bool) {
	body, ok := ctx.Value(rawBodyContextKey{}).([]byte)
	return body, ok
}

// messagesJSON returns the request's message objects as JSON.
// The generated messagingapi.Message keeps only the fields shared by every message type,
// so the objects are taken from the raw body to preserve text, URLs, and other type-specific fields.
// It falls back to the typed messages when the raw body isn't available.
func messagesJSON(ctx context.Context, messages []messagingapi.Message) ([]byte, error) {
	rawBody, ok := rawRequestBody(ctx)
	if !ok {
		return json.Marshal(messages)
	}
	var body struct {
		Messages json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(rawBody, &body); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}
	if len(body.Messages) == 0 {
		return json.Marshal(messages)
	}
	return body.Messages, nil
}
