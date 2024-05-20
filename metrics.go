package main

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
)

func init() {
	metrics.GetOrCreateCounter("cloudflare_requests_total")
	metrics.GetOrCreateHistogram("cloudflare_scrape_duration_seconds")
}

func generateMetrics(data interface{}) {
	// Example implementation: Iterate through the data and create appropriate metrics
	// This is a placeholder; adjust the implementation according to the actual data structure
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			switch val := value.(type) {
			case float64:
				metrics.GetOrCreateGauge(fmt.Sprintf("cloudflare_%s", key), func() float64 { return val }).Set(val)
			case int64:
				metrics.GetOrCreateGauge(fmt.Sprintf("cloudflare_%s", key), func() float64 { return float64(val) }).Set(float64(val))
			case string:
				// Handle string values if needed
			// Add more cases as needed to handle different data types
			default:
				// Handle other types if necessary
			}
		}
	default:
		// Handle other data structures if necessary
	}
}
