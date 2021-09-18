package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	bulkheadFullCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkhead_full_exception_count",
		Help: "The total number of bulkhead full exception events",
	}, []string{"bulkhead"})
	bulkheadWaitMSSum = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkhead_incr_wait_sum_ms",
		Help: "The sum of time spent in bulkhead Incr in ms",
	}, []string{"bulkhead"})
)

func IncrBulkheadFull(name string) {
	bulkheadFullCount.WithLabelValues(name).Inc()
}

func IncrBulkheadWaitSum(name string, t time.Duration) {
	bulkheadWaitMSSum.WithLabelValues(name).Add(float64(t.Milliseconds()))
}
