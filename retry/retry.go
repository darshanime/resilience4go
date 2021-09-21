package retry

import (
	"net/http"
	"sync"
	"time"

	"github.com/darshanime/resilience4go/metrics"
)

type Retry struct {
	maxAttempts    int
	retryMap       map[*http.Request]int
	backoff        func() time.Duration
	retryPredicate func(req *http.Request, resp *http.Response, err error) bool
	reqNamer       func(req *http.Request) string
	mu             sync.Mutex
}

func defaultRetryBackoff() time.Duration {
	return time.Duration(0)
}

func defaultRequestNamer(req *http.Request) string {
	return req.URL.String()
}

func New(maxAttempts int) *Retry {
	return &Retry{
		maxAttempts:    maxAttempts,
		retryMap:       make(map[*http.Request]int),
		backoff:        defaultRetryBackoff,
		retryPredicate: OnServerErrors,
		reqNamer:       defaultRequestNamer,
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

func (r *Retry) WithRequestNamer(reqNamer func(req *http.Request) string) *Retry {
	r.reqNamer = reqNamer
	return r
}

func (r *Retry) Wait(req *http.Request) {
	time.Sleep(r.backoff())

	r.mu.Lock()
	r.retryMap[req]++
	r.mu.Unlock()
}

func (r *Retry) ShouldRetry(req *http.Request, resp *http.Response, err error) bool {
	if r.retryPredicate(req, resp, err) {
		r.mu.Lock()
		defer r.mu.Unlock()

		if r.retryMap[req] < r.maxAttempts {
			return true
		}
		delete(r.retryMap, req)
		return false
	}
	return false
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
