package webhook

import (
	"context"
	"sync"
)

// MockClient is a mock implementation of the webhook Client interface for testing
type MockClient struct {
	SendEventsFunc func(ctx context.Context, req *SendEventsRequest) error

	mu    sync.Mutex
	calls []*SendEventsRequest
}

// SendEvents implements Client.SendEvents for testing
func (m *MockClient) SendEvents(ctx context.Context, req *SendEventsRequest) error {
	m.mu.Lock()
	m.calls = append(m.calls, req)
	m.mu.Unlock()
	if m.SendEventsFunc != nil {
		return m.SendEventsFunc(ctx, req)
	}
	return nil
}

// SendEventsCalls returns the requests received so far
func (m *MockClient) SendEventsCalls() []*SendEventsRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]*SendEventsRequest(nil), m.calls...)
}

// NewTestClient creates a new webhook client for testing
func NewTestClient() Client {
	return NewClient()
}

// NewMockClient creates a new mock webhook client for testing
func NewMockClient() *MockClient {
	return &MockClient{}
}
