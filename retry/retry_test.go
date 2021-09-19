package retry

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryNew(t *testing.T) {
	maxAttempts := 5
	r := New(maxAttempts)

	assert.Equal(t, maxAttempts, r.maxAttempts)
	assert.NotNil(t, r.retryPredicate)
	assert.NotNil(t, r.backoff)

	constantBackoff := ConstantBackoff(time.Duration(1 * time.Second))
	r = r.WithBackoffFunction(constantBackoff)
	assert.NotNil(t, r.backoff)

	rp := func(req *http.Request, resp *http.Response, err error) bool { return false }
	r = r.WithRetryPredicate(rp)
	assert.NotNil(t, r.backoff)

}

func TestConstantBackoff(t *testing.T) {
	constant := ConstantBackoff(time.Duration(10 * time.Second))
	expectedBackoffs := []time.Duration{
		10 * time.Second,
		10 * time.Second,
		10 * time.Second,
	}
	for _, d := range expectedBackoffs {
		assert.Equal(t, d, constant())
	}
}

func TestLinearBackoff(t *testing.T) {
	linear := LinearBackoff(time.Duration(10*time.Second), time.Duration(1*time.Minute))
	expectedBackoffs := []time.Duration{
		10 * time.Second,
		20 * time.Second,
		30 * time.Second,
		40 * time.Second,
		50 * time.Second,
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
	}
	for _, d := range expectedBackoffs {
		assert.Equal(t, d, linear())
	}
}

func TestExponentialBackoff(t *testing.T) {
	exponential := ExponentialBackoff(time.Duration(5*time.Second), time.Duration(1*time.Minute))
	expectedBackoffs := []time.Duration{
		5 * time.Second,
		10 * time.Second,
		20 * time.Second,
		35 * time.Second,
		55 * time.Second,
		60 * time.Second,
		60 * time.Second,
		60 * time.Second,
	}
	for _, d := range expectedBackoffs {
		assert.Equal(t, d, exponential())
	}
}
