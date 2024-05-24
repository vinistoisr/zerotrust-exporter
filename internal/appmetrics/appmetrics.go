package appmetrics

import (
	"github.com/VictoriaMetrics/metrics"
)

// Prometheus Endpoint metrics
var (
	UpMetric         = metrics.NewGauge("zerotrust_exporter_up", func() float64 { return 1 })
	ScrapeDuration   = metrics.NewHistogram("zerotrust_exporter_scrape_duration_seconds")
	ApiCallCounter   = metrics.NewCounter("zerotrust_exporter_api_calls_total")
	ApiErrorsCounter = metrics.NewCounter("zerotrust_exporter_api_errors_total")
)

func SetUpMetric(value float64) {
	UpMetric.Set(value)
}

func SetScrapeDuration(value float64) {
	ScrapeDuration.Update(value)
}

func IncApiCallCounter() {
	ApiCallCounter.Inc()
}

func IncApiErrorsCounter() {
	ApiErrorsCounter.Inc()
}
