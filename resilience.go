package resilience

import (
	"net/http"
	"time"

	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/darshanime/resilience4go/retry"
)

type Resilience struct {
	name    string
	next    http.RoundTripper
	timeout time.Duration

	bh    *bulkhead.Bulkhead
	retry *retry.Retry
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

	resp, err := r.next.RoundTrip(req)
	closeResp(resp)

	for r.retry.ShouldRetry(req, resp, err) {
		r.retry.Wait(req)
		resp, err = r.next.RoundTrip(req)
		closeResp(resp)
	}
	return resp, err
}

func New(name string) *Resilience {
	return &Resilience{
		name: name,
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
	return hc
}

func (r *Resilience) Build() *http.Client {
	return r.BuildWithHTTPClient(http.DefaultClient)
}
