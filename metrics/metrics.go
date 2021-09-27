package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	model "github.com/prometheus/client_model/go"
)

type Metrics struct {
	registry      prometheus.Registerer
	buckets       []float64
	mu            sync.Mutex
	collectorChan chan prometheus.Metric
}

var (
	BulkheadFullCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkhead_full_exception_count",
		Help: "The total number of bulkhead full exception events",
	}, []string{"bulkhead"})
	BulkheadWaitMSSum = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkhead_incr_wait_sum_ms",
		Help: "The sum of time spent in bulkhead Incr in ms",
	}, []string{"bulkhead"})
	BulkheadBufferLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bulkhead_buffer_length",
		Help: "The number of requests that are in the bulkhead buffer",
	}, []string{"bulkhead"})
	BulkheadMaxBufferLength = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bulkhead_max_buffer_length",
		Help: "The max number of requests that can be in the bulkhead buffer",
	}, []string{"bulkhead"})
	RetryCountTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "retry_count_total",
		Help: "The total retry count of requests",
	}, []string{"request"})
	HTTPResponseCode = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_response_codes_total",
		Help: "The total count of response codes for requests",
	}, []string{"request", "code"})
	HTTPResponseSize = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_response_size_total",
		Help: "The total size of response for requests",
	}, []string{"request", "code"})
)

func IncrHTTPResponseCode(name, code string) {
	HTTPResponseCode.WithLabelValues(name, code).Inc()
}

func IncrHTTPResponseSize(name, code string, value float64) {
	HTTPResponseSize.WithLabelValues(name, code).Add(value)
}

func ObserveHTTPResponseLatency(name, code string, value float64) {
	HTTPResponseSize.WithLabelValues(name, code).Add(value)
}

func IncrBulkheadFull(name string) {
	BulkheadFullCount.WithLabelValues(name).Inc()
}

func IncrBulkheadWaitSum(name string, t time.Duration) {
	BulkheadWaitMSSum.WithLabelValues(name).Add(float64(t.Milliseconds()))
}

func SetBulkheadBufferLength(name string, value float64) {
	BulkheadBufferLength.WithLabelValues(name).Set(value)
}

func SetBulkheadMaxBufferLength(name string, value float64) {
	BulkheadMaxBufferLength.WithLabelValues(name).Set(value)
}

func IncrRetryCountTotal(name string) {
	RetryCountTotal.WithLabelValues(name).Inc()
}

func New() *Metrics {
	return &Metrics{
		registry:      prometheus.DefaultRegisterer,
		buckets:       prometheus.DefBuckets,
		collectorChan: make(chan prometheus.Metric, 1),
	}
}

func (m *Metrics) WithRegisterer(r prometheus.Registerer) *Metrics {
	m.registry = r
	return m
}

func (m *Metrics) WithBuckets(b []float64) *Metrics {
	m.buckets = b
	return m
}

func (m *Metrics) Build() *Metrics {
	m.registry.MustRegister(
		BulkheadFullCount,
		BulkheadWaitMSSum,
		BulkheadBufferLength,
		BulkheadMaxBufferLength,
		RetryCountTotal)
	return m
}

func (m *Metrics) GetCounterValue(col prometheus.Collector) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	col.Collect(m.collectorChan)
	metric := model.Metric{}
	if err := (<-m.collectorChan).Write(&metric); err != nil {
		return 0.0, err
	}
	return *metric.Counter.Value, nil
}

func (m *Metrics) GetGaugeValue(col prometheus.Collector) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	col.Collect(m.collectorChan)
	metric := model.Metric{}
	if err := (<-m.collectorChan).Write(&metric); err != nil {
		return 0.0, err
	}
	return *metric.Gauge.Value, nil
}
