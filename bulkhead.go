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

type bulkhead struct {
	domain          string
	maxWaitDuration time.Duration
	activeCount     int
	mu              sync.Mutex
	buffer          chan struct{}
}

func New(domain string) *bulkhead {
	return &bulkhead{
		domain:          domain,
		maxWaitDuration: maxWaitDuration,
		activeCount:     0,
		buffer:          make(chan struct{}, maxConcurrentCalls),
	}
}

func (b *bulkhead) WithMaxParallelCalls(calls int) *bulkhead {
	b.buffer = make(chan struct{}, calls)
	return b
}

func (b *bulkhead) WithMaxWaitDuration(t time.Duration) *bulkhead {
	b.maxWaitDuration = t
	return b
}

func (b *bulkhead) Incr() error {
	select {
	case <-time.After(b.maxWaitDuration):
		return errors.New(BulkHeadFullError)
	case b.buffer <- struct{}{}:
		b.activeCount++
		return nil
	}
}

func (b *bulkhead) Decr() {
	b.mu.Lock()
	if b.activeCount > 0 {
		b.activeCount--
		<-b.buffer
	}
	b.mu.Unlock()
}
