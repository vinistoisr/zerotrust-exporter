package collector

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/appmetrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/config"
	"github.com/vinistoisr/zerotrust-exporter/internal/devices"
	"github.com/vinistoisr/zerotrust-exporter/internal/dex"
	"github.com/vinistoisr/zerotrust-exporter/internal/tunnels"
	"github.com/vinistoisr/zerotrust-exporter/internal/users"
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
	// create a channel between device metrics and user metrics
	deviceMetricsChan := make(chan map[string]devices.DeviceStatus)

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(4)

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

	// Go Collect dex metrics
	go func() {
		defer wg.Done()
		if config.EnableDex {
			log.Println("Collecting dex metrics...")
			dex.CollectDexMetrics(context.TODO(), config.AccountID)
		}
	}()

	// Wait for all metrics collection to complete
	log.Println("Waiting for all metrics collection to complete...")
	wg.Wait()
	log.Println("All metrics collection completed.")

	// Update scrape duration metric
	appmetrics.ScrapeDuration.UpdateDuration(startTime)
	// Write metrics to the response
	metrics.WritePrometheus(w, true)

	// Print debug information if enabled
	if config.Debug {
		log.Printf("Scrape completed in %v", time.Since(startTime))
		log.Printf("API calls made: %d", appmetrics.ApiCallCounter.Get())
		log.Printf("API errors encountered: %d", appmetrics.ApiErrorsCounter.Get())
	}
}
