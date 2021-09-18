package bulkhead

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBulkheadNew(t *testing.T) {
	domain := "foo.bar"
	bh := New(domain)

	assert.Equal(t, domain, bh.domain)

	maxParallelCalls := 1
	bh = bh.WithMaxParallelCalls(maxParallelCalls)
	assert.Zero(t, len(bh.buffer))
	assert.Zero(t, bh.activeCount)
}

func TestBulkheadIncr(t *testing.T) {
	bh := New("foo.bar").WithMaxParallelCalls(1)

	err := bh.Incr()
	assert.Equal(t, 1, len(bh.buffer))
	assert.Nil(t, err)

	err = bh.Incr()
	assert.EqualError(t, err, BulkHeadFullError)
}
