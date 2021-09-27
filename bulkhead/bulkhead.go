package bulkhead

import (
	"sync"
	"time"

	"github.com/darshanime/resilience4go/metrics"
)

const (
	maxConcurrentCalls = 10
	maxWaitDuration    = 500 * time.Microsecond
	defaultName        = "default"
)

type Bulkhead struct {
	name            string
	maxWaitDuration time.Duration
	active          bool
	m               *metrics.Metrics

	buffer chan struct{}
	mu     sync.Mutex
}

func New() *Bulkhead {
	return &Bulkhead{
		name:            defaultName,
		maxWaitDuration: maxWaitDuration,
		buffer:          make(chan struct{}, maxConcurrentCalls),
		active:          true,
	}
}

func (b *Bulkhead) WithMaxParallelCalls(calls int) *Bulkhead {
	b.buffer = make(chan struct{}, calls)
	// TODO: if WithName is called after WithMaxParallelCalls, it may add a spurious metric with bulkhead=default
	metrics.SetBulkheadMaxBufferLength(b.name, float64(calls))
	return b
}

func (b *Bulkhead) WithName(name string) *Bulkhead {
	b.name = name
	return b
}

func (b *Bulkhead) WithMetrics(m *metrics.Metrics) *Bulkhead {
	b.m = m
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

func (b *Bulkhead) ResizeBulkhead(newLen int) *Bulkhead {
	if newLen == 0 || newLen == len(b.buffer) {
		return b
	}

	metrics.SetBulkheadMaxBufferLength(b.name, float64(newLen))

	b.mu.Lock()
	defer b.mu.Unlock()

	activeLen := len(b.buffer)
	newChan := make(chan struct{}, newLen)
	b.buffer = newChan

	if newLen <= activeLen {
		return b
	}
	for i := 0; i < activeLen; i++ {
		b.buffer <- struct{}{}
	}

	return b
}

func (b *Bulkhead) Incr() error {
	if b == nil || !b.active {
		return nil
	}

	start := time.Now()
	defer func() {
		metrics.IncrBulkheadWaitSum(b.name, time.Since(start))
		metrics.SetBulkheadBufferLength(b.name, float64(len(b.buffer)))
	}()

	select {
	case <-time.After(b.maxWaitDuration):
		metrics.IncrBulkheadFull(b.name)
		return ErrFull
	case b.buffer <- struct{}{}:
		return nil
	}
}

func (b *Bulkhead) Decr() {
	if b == nil || !b.active {
		return
	}

	b.mu.Lock()
	if len(b.buffer) > 0 {
		<-b.buffer
	}
	b.mu.Unlock()

	metrics.SetBulkheadBufferLength(b.name, float64(len(b.buffer)))
}

func (b *Bulkhead) GetBulkheadFullCount() (float64, error) {
	counter, err := metrics.BulkheadFullCount.GetMetricWithLabelValues(b.name)
	if err != nil {
		return 0.0, err
	}

	return b.m.GetCounterValue(counter)
}
