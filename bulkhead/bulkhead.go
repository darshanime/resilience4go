package bulkhead

import (
	"errors"
	"sync"
	"time"

	"github.com/darshanime/resilience4go/metrics"
)

const (
	maxConcurrentCalls = 10
	maxWaitDuration    = 500 * time.Microsecond
)

type Bulkhead struct {
	name            string
	maxWaitDuration time.Duration
	activeCount     int
	mu              sync.Mutex
	buffer          chan struct{}
	active          bool
}

func New() *Bulkhead {
	return &Bulkhead{
		name:            "default",
		maxWaitDuration: maxWaitDuration,
		activeCount:     0,
		buffer:          make(chan struct{}, maxConcurrentCalls),
		active:          true,
	}
}

func (b *Bulkhead) WithMaxParallelCalls(calls int) *Bulkhead {
	b.buffer = make(chan struct{}, calls)
	return b
}

func (b *Bulkhead) WithName(name string) *Bulkhead {
	b.name = name
	return b
}

func (b *Bulkhead) WithMaxWaitDuration(t time.Duration) *Bulkhead {
	b.maxWaitDuration = t
	return b
}

func (b *Bulkhead) DisableBulkhead() *Bulkhead {
	b.active = false
	return b
}

func (b *Bulkhead) Incr() error {
	if b == nil || !b.active {
		return nil
	}

	start := time.Now()
	defer func() {
		metrics.IncrBulkheadWaitSum(b.name, time.Since(start))
	}()

	select {
	case <-time.After(b.maxWaitDuration):
		metrics.IncrBulkheadFull(b.name)
		return errors.New(BulkHeadFullError)
	case b.buffer <- struct{}{}:
		b.activeCount++
		return nil
	}
}

func (b *Bulkhead) Decr() {
	if b == nil || !b.active {
		return
	}

	b.mu.Lock()
	if b.activeCount > 0 {
		b.activeCount--
		<-b.buffer
	}
	b.mu.Unlock()
}
