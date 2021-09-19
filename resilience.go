package resilience

import (
	"net/http"
	"time"

	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/darshanime/resilience4go/retry"
)

type resilience struct {
	name    string
	next    http.RoundTripper
	timeout time.Duration

	bh    *bulkhead.Bulkhead
	retry *retry.Retry
}

func (r *resilience) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := r.bh.Incr(); err != nil {
		return nil, err
	}
	defer r.bh.Decr()

	resp, err := r.next.RoundTrip(req)

	for r.retry.ShouldRetry(req, resp, err) && r.retry.Retry(req) {
		resp, err = r.next.RoundTrip(req)
	}
	return resp, err
}

func New(name string) *resilience {
	return &resilience{
		name: name,
	}
}

func (r *resilience) WithBulkHead(bh *bulkhead.Bulkhead) *resilience {
	r.bh = bh.WithName(r.name)
	return r
}

func (r *resilience) WithRetry(rt *retry.Retry) *resilience {
	r.retry = rt
	return r
}

// WithRequestTimeout will set the httpClient.Timeout to passed value
func (r *resilience) WithRequestTimeout(timeout time.Duration) *resilience {
	r.timeout = timeout
	return r
}

func (r *resilience) BuildWithHTTPClient(hc *http.Client) *http.Client {
	if hc.Transport == nil {
		hc.Transport = http.DefaultTransport
	}

	if r.timeout != 0 {
		hc.Timeout = r.timeout
	}

	r.next = hc.Transport
	hc.Transport = r
	return hc
}

func (r *resilience) Build() *http.Client {
	return r.BuildWithHTTPClient(http.DefaultClient)
}
