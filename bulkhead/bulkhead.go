package bulkhead

import (
	"errors"
	"sync"
	"time"
)

const (
	maxConcurrentCalls = 10
	maxWaitDuration    = 500 * time.Microsecond
)

type Bulkhead struct {
	maxWaitDuration time.Duration
	activeCount     int
	mu              sync.Mutex
	buffer          chan struct{}
	active          bool
}

func New() *Bulkhead {
	return &Bulkhead{
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

	select {
	case <-time.After(b.maxWaitDuration):
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
