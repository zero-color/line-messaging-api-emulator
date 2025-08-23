package shortid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Parallel()

	got := New()
	assert.Len(t, got, 11)
}
