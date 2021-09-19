package retry

import (
	"net/http"
	"sync"
	"time"
)

type Retry struct {
	maxAttempts    int
	retryMap       map[*http.Request]int
	backoff        func() time.Duration
	retryPredicate func(req *http.Request, resp *http.Response, err error) bool
	mu             sync.Mutex
}

func defaultRetryBackoff() time.Duration {
	return time.Duration(0)
}

func New(maxAttempts int) *Retry {
	return &Retry{
		maxAttempts:    maxAttempts,
		retryMap:       make(map[*http.Request]int),
		backoff:        defaultRetryBackoff,
		retryPredicate: OnServerErrors,
	}
}

func (r *Retry) WithBackoffFunction(f func() time.Duration) *Retry {
	r.backoff = f
	return r
}

func (r *Retry) WithRetryPredicate(retryPredicate func(req *http.Request, resp *http.Response, err error) bool) *Retry {
	r.retryPredicate = retryPredicate
	return r
}

func (r *Retry) Retry(req *http.Request) bool {
	r.mu.Lock()
	if r.retryMap[req] < r.maxAttempts {
		r.retryMap[req]++
		r.mu.Unlock()
		time.Sleep(r.backoff())
		return true
	}
	delete(r.retryMap, req)
	r.mu.Unlock()
	return false
}

func (r *Retry) ShouldRetry(req *http.Request, resp *http.Response, err error) bool {
	return r.retryPredicate(req, resp, err)
}

func OnErrors(req *http.Request, resp *http.Response, err error) bool {
	return err != nil
}

func OnServerErrors(req *http.Request, resp *http.Response, err error) bool {
	return err != nil || (resp.StatusCode >= 500 && resp.StatusCode <= 599)
}

func ConstantBackoff(interval time.Duration) func() time.Duration {
	return func() time.Duration {
		return interval
	}
}

func LinearBackoff(interval time.Duration, maxWait time.Duration) func() time.Duration {
	lastWait := time.Duration(0)
	return func() time.Duration {
		proposedWait := lastWait + interval
		if proposedWait > maxWait {
			lastWait = maxWait
			return maxWait
		}
		lastWait = proposedWait
		return proposedWait
	}
}

func ExponentialBackoff(exponent time.Duration, maxWait time.Duration) func() time.Duration {
	lastWait := exponent
	iteration := float64(0)
	return func() time.Duration {
		proposedWait := lastWait + time.Duration(iteration*float64(exponent))
		iteration++
		if proposedWait > maxWait {
			lastWait = maxWait
			return maxWait
		}
		lastWait = proposedWait
		return proposedWait
	}
}
