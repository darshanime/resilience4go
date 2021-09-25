package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	resilience "github.com/darshanime/resilience4go"
	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/darshanime/resilience4go/retry"
)

type loggerTripper struct {
	next http.RoundTripper
}

func (l *loggerTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("%s: calling\n", req.URL)
	resp, err := l.next.RoundTrip(req)
	fmt.Printf("%s: resp %+v, error %s\n", req.URL, resp, err)
	return resp, err
}

func main() {
	bh := bulkhead.New().WithMaxParallelCalls(2).WithMaxWaitDuration(1 * time.Second)

	rt := retry.New(3).WithBackoffFunction(
		retry.ConstantBackoff(2 * time.Second),
	)

	clientPlus := resilience.New("user_service").WithBulkHead(bh).WithRetry(rt).WithRequestTimeout(
		100 * time.Millisecond,
	).Build()

	clientPlus.Get("http://google.com")

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		i := i
		wg.Add(1)
		go func() {
			url := "https://httpbin.org/status/" + fmt.Sprintf("%d", 500+i)
			clientPlus.Get(url)
			wg.Done()
		}()
	}
	wg.Wait()
}
