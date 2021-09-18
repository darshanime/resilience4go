package resilience

import (
	"net/http"

	"github.com/darshanime/resilience4go/bulkhead"
)

type resilience struct {
	name string
	next http.RoundTripper

	bh *bulkhead.Bulkhead
}

func (r *resilience) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := r.bh.Incr(); err != nil {
		return nil, err
	}
	defer r.bh.Decr()
	return r.next.RoundTrip(req)
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

func (r *resilience) BuildWithHTTPClient(hc *http.Client) *http.Client {
	if hc.Transport == nil {
		hc.Transport = http.DefaultTransport
	}

	r.next = hc.Transport
	hc.Transport = r
	return hc
}

func (r *resilience) Build() *http.Client {
	return r.BuildWithHTTPClient(http.DefaultClient)
}
