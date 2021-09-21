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
	bulkheadBufferLength = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bulkhead_buffer_length",
		Help: "The number of requests that are in the bulkhead buffer",
	}, []string{"bulkhead"})
	retryCountTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "retry_count_total",
		Help: "The total retry count of requests",
	}, []string{"request"})
)

func IncrBulkheadFull(name string) {
	bulkheadFullCount.WithLabelValues(name).Inc()
}

func IncrBulkheadWaitSum(name string, t time.Duration) {
	bulkheadWaitMSSum.WithLabelValues(name).Add(float64(t.Milliseconds()))
}

func SetBulkheadBufferLength(name string, value float64) {
	bulkheadBufferLength.WithLabelValues(name).Set(value)
}

func IncrRetryCountTotal(name string) {
	retryCountTotal.WithLabelValues(name).Inc()
}
