package resilience

import (
	"net/http"
	"testing"

	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/stretchr/testify/assert"
)

type noopRoundTripper struct {
	next http.RoundTripper
}

func (n *noopRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return n.next.RoundTrip(req)
}

func TestResilience(t *testing.T) {
	r := New("test")

	assert.Equal(t, "test", r.name)

	bh := bulkhead.New()
	r = r.WithBulkHead(bh)
	assert.Equal(t, bh, r.bh)

	r.Build()
	assert.Equal(t, http.DefaultTransport, r.next)

	myHTTPClient := &http.Client{}
	myTransport := &noopRoundTripper{next: http.DefaultTransport}
	myHTTPClient.Transport = myTransport

	newHTTPClient := r.BuildWithHTTPClient(myHTTPClient)
	assert.Equal(t, myTransport, r.next)
	assert.Equal(t, r, newHTTPClient.Transport)
}
