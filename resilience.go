package resilience

import (
	"fmt"
	"net/http"
	"time"

	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/darshanime/resilience4go/metrics"
	"github.com/darshanime/resilience4go/retry"
)

type Resilience struct {
	name     string
	reqNamer func(req *http.Request) string
	next     http.RoundTripper
	timeout  time.Duration

	bh    *bulkhead.Bulkhead
	retry *retry.Retry
	m     *metrics.Metrics
}

func defaultRequestNamer(req *http.Request) string {
	return req.URL.String()
}

func closeResp(resp *http.Response) {
	if resp != nil {
		resp.Body.Close()
	}
}

func (r *Resilience) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := r.bh.Incr(); err != nil {
		return nil, err
	}
	defer r.bh.Decr()

	reqName := r.reqNamer(req)
	resp, err := r.next.RoundTrip(req)
	if r.m != nil && resp != nil {
		metrics.IncrHTTPResponseCode(reqName, fmt.Sprintf("%d", resp.StatusCode))
	}
	closeResp(resp)

	for r.retry.ShouldRetry(req, resp, reqName, err) {
		r.retry.Wait(req)
		resp, err = r.next.RoundTrip(req)
		closeResp(resp)
	}
	return resp, err
}

func (r *Resilience) WithRequestNamer(reqNamer func(req *http.Request) string) *Resilience {
	r.reqNamer = reqNamer
	return r
}

func New(name string) *Resilience {
	return &Resilience{
		name:     name,
		reqNamer: defaultRequestNamer,
	}
}

func (r *Resilience) WithBulkHead(bh *bulkhead.Bulkhead) *Resilience {
	r.bh = bh.WithName(r.name)
	return r
}

func (r *Resilience) WithRetry(rt *retry.Retry) *Resilience {
	r.retry = rt
	return r
}

// WithRequestTimeout will set the httpClient.Timeout to passed value
func (r *Resilience) WithRequestTimeout(timeout time.Duration) *Resilience {
	r.timeout = timeout
	return r
}

func (r *Resilience) WithMetrics(m *metrics.Metrics) *Resilience {
	r.m = m
	return r
}

// BuildWithHTTPClient will accept a http client from user
func (r *Resilience) BuildWithHTTPClient(hc *http.Client) *http.Client {
	if hc.Transport == nil {
		hc.Transport = http.DefaultTransport
	}

	if r.timeout != 0 {
		hc.Timeout = r.timeout
	}

	r.next = hc.Transport
	hc.Transport = r

	// passing metrics
	r.bh = r.bh.WithMetrics(r.m)
	return hc
}

func (r *Resilience) Build() *http.Client {
	return r.BuildWithHTTPClient(http.DefaultClient)
}
