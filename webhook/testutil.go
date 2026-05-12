package webhook

import "context"

// MockClient is a mock implementation of the webhook Client interface for testing
type MockClient struct {
	SendFollowEventsFunc  func(ctx context.Context, req *FollowEventRequest) error
	SendFollowEventsCalls []*FollowEventRequest
}

// SendFollowEvents implements Client.SendFollowEvents for testing
func (m *MockClient) SendFollowEvents(ctx context.Context, req *FollowEventRequest) error {
	m.SendFollowEventsCalls = append(m.SendFollowEventsCalls, req)
	if m.SendFollowEventsFunc != nil {
		return m.SendFollowEventsFunc(ctx, req)
	}
	return nil
}

// NewTestClient creates a new webhook client for testing
func NewTestClient() Client {
	return NewClient()
}

// NewMockClient creates a new mock webhook client for testing
func NewMockClient() *MockClient {
	return &MockClient{
		SendFollowEventsCalls: []*FollowEventRequest{},
	}
}
