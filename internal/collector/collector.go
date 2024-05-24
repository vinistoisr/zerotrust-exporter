package collector

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vinistoisr/zerotrust_exporter/internal/config"
	"github.com/vinistoisr/zerotrust_exporter/internal/devices"
	"github.com/vinistoisr/zerotrust_exporter/internal/tunnels"
	"github.com/vinistoisr/zerotrust_exporter/internal/users"
)

// Prometheus Endpoint metrics
var (
	UpMetric         = metrics.NewGauge("zerotrust_exporter_up", func() float64 { return 1 })
	ScrapeDuration   = metrics.NewHistogram("zerotrust_exporter_scrape_duration_seconds")
	ApiCallCounter   = metrics.NewCounter("zerotrust_exporter_api_calls_total")
	ApiErrorsCounter = metrics.NewCounter("zerotrust_exporter_api_errors_total")
)

// Register metrics handler
func RegisterHandler() {
	http.HandleFunc("/metrics", MetricsHandler)
}

// StartServer starts the HTTP server
func StartServer(addr string) {
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// metricsHandler handles the /metrics endpoint
func MetricsHandler(w http.ResponseWriter, req *http.Request) {
	// Start timer for scrape duration
	startTime := time.Now()

	// Using goroutines to collect metrics for devices, users, and tunnels concurrently
	deviceMetricsChan := make(chan map[string]devices.DeviceStatus)

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(3)

	// GO Collect device metrics
	go func() {
		defer wg.Done()
		if config.EnableDevices {
			log.Println("Collecting device metrics...")
			deviceMetrics := devices.CollectDeviceMetrics()
			deviceMetricsChan <- deviceMetrics
			close(deviceMetricsChan)
		} else {
			close(deviceMetricsChan)
		}
	}()

	// GO Collect user metrics
	go func() {
		defer wg.Done()
		if config.EnableUsers {
			log.Println("Waiting for device metrics...")
			deviceMetrics, ok := <-deviceMetricsChan
			if ok {
				users.CollectUserMetrics(deviceMetrics)
			} else {
				log.Println("Failed to read device metrics from channel.")
			}
		}
	}()

	// GO Collect tunnel metrics
	go func() {
		defer wg.Done()
		if config.EnableTunnels {
			log.Println("Collecting tunnel metrics...")
			tunnels.CollectTunnelMetrics()
		}
	}()

	// Wait for all metrics collection to complete
	log.Println("Waiting for all metrics collection to complete...")
	wg.Wait()
	log.Println("All metrics collection completed.")

	// Update scrape duration metric
	ScrapeDuration.UpdateDuration(startTime)
	// Write metrics to the response
	metrics.WritePrometheus(w, true)

	// Print debug information if enabled
	if config.Debug {
		log.Printf("Scrape completed in %v", time.Since(startTime))
		log.Printf("API calls made: %d", ApiCallCounter.Get())
		log.Printf("API errors encountered: %d", ApiErrorsCounter.Get())
	}
}
