package main

import (
	"fmt"
	"time"

	resilience "github.com/darshanime/resilience4go"
	"github.com/darshanime/resilience4go/bulkhead"
	"github.com/darshanime/resilience4go/metrics"
	"github.com/darshanime/resilience4go/retry"
)

func main() {
	bulkheadSize := 5
	bh := bulkhead.New().WithName("user_service").WithMaxParallelCalls(bulkheadSize).WithMaxWaitDuration(1 * time.Second)

	rt := retry.New(3).WithBackoffFunction(
		retry.ConstantBackoff(2 * time.Second),
	)

	m := metrics.New().Build()

	clientPlus := resilience.New("user_service").WithBulkHead(bh).WithRetry(rt).WithRequestTimeout(
		100 * time.Millisecond,
	).WithMetrics(m).Build()

	go func() {
		ticker := time.NewTicker(50 * time.Second)
		oldValue := float64(0)

		for {
			<-ticker.C

			value, _ := bh.GetBulkheadFullCount()
			if value > oldValue {
				oldValue = value
				bulkheadSize *= 2
				bh.ResizeBulkhead(bulkheadSize)
			}
		}
	}()

	for i := 0; i < 5; i++ {
		url := "https://httpbin.org/status/" + fmt.Sprintf("%d", 500+i)
		clientPlus.Get(url)
	}
}
