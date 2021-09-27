package main

import (
	"fmt"
	"net/http"
	"time"

	resilience "github.com/darshanime/resilience4go"
	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/darshanime/resilience4go/metrics"
	"github.com/darshanime/resilience4go/retry"
)

type loggerTripper struct {
	next http.RoundTripper
}

func main() {
	bh := bulkhead.New().WithMaxParallelCalls(2).WithMaxWaitDuration(1 * time.Second)

	rt := retry.New(3).WithBackoffFunction(
		retry.ConstantBackoff(2 * time.Second),
	)

	m := metrics.New().Build()

	clientPlus := resilience.New("user_service").WithBulkHead(bh).WithRetry(rt).WithRequestTimeout(
		100 * time.Millisecond,
	).WithMetrics(m).Build()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		oldValue := float64(0)

		for {
			<-ticker.C

			value, _ := bh.GetBulkheadFullCount()
			if value > oldValue {
				oldValue = value
				bh.ResizeBulkhead(int(value))
			}
		}
	}()

	for i := 0; i < 5; i++ {
		url := "https://httpbin.org/status/" + fmt.Sprintf("%d", 500+i)
		clientPlus.Get(url)
	}
}
