package bulkhead

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBulkheadNew(t *testing.T) {
	bh := New()
	maxParallelCalls := 1
	bh = bh.WithMaxParallelCalls(maxParallelCalls)
	assert.Zero(t, len(bh.buffer))
	assert.Zero(t, bh.activeCount)

	assert.True(t, bh.active)
	bh = bh.DisableBulkhead()
	assert.False(t, bh.active)
}

func TestBulkheadIncr(t *testing.T) {
	bh := New().WithMaxParallelCalls(1)

	err := bh.Incr()
	assert.Equal(t, 1, len(bh.buffer))
	assert.Nil(t, err)

	err = bh.Incr()
	assert.EqualError(t, err, BulkHeadFullError)
}
