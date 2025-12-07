package db

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPMiddleware(t *testing.T) {
	t.Parallel()

	db := NewTestDB(t)
	req := (&http.Request{}).WithContext(context.Background())
	HTTPMiddleware(db)(testNextHandler{t: t, expectedDB: db}).ServeHTTP(nil, req)
}

type testNextHandler struct {
	t          *testing.T
	expectedDB DB
}

func (t testNextHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	assert.Equal(t.t, t.expectedDB, FromContext(r.Context()))
}
