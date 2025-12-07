package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	t.Parallel()

	db := NewTestDB(t)
	ctx := ToContext(context.Background(), db)
	assert.Equal(t, db, FromContext(ctx))
}
