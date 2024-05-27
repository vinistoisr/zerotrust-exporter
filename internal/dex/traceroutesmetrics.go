package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/appmetrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/config"
)

// TracerouteStats represents the detailed stats for a traceroute test
type TracerouteStats struct {
	UniqueDevicesTotal int `json:"uniqueDevicesTotal"`
	RoundTripTimeMs    struct {
		Min   float64 `json:"min"`
		Avg   float64 `json:"avg"`
		Max   float64 `json:"max"`
		Slots []struct {
			Timestamp string  `json:"timestamp"`
			Value     float64 `json:"value"`
		} `json:"slots"`
	} `json:"roundTripTimeMs"`
	HopsCount struct {
		Min   float64 `json:"min"`
		Avg   float64 `json:"avg"`
		Max   float64 `json:"max"`
		Slots []struct {
			Timestamp string  `json:"timestamp"`
			Value     float64 `json:"value"`
		} `json:"slots"`
	} `json:"hopsCount"`
	PacketLossPct struct {
		Min   float64 `json:"min"`
		Avg   float64 `json:"avg"`
		Max   float64 `json:"max"`
		Slots []struct {
			Timestamp string  `json:"timestamp"`
			Value     float64 `json:"value"`
		} `json:"slots"`
	} `json:"packetLossPct"`
	AvailabilityPct struct {
		Min   float64 `json:"min"`
		Avg   float64 `json:"avg"`
		Max   float64 `json:"max"`
		Slots []struct {
			Timestamp string  `json:"timestamp"`
			Value     float64 `json:"value"`
		} `json:"slots"`
	} `json:"availabilityPct"`
}

// TracerouteTestResponse represents the response for a traceroute test
type TracerouteTestResponse struct {
	Result   TracerouteTestResult `json:"result"`
	Success  bool                 `json:"success"`
	Errors   []interface{}        `json:"errors"`
	Messages []interface{}        `json:"messages"`
}

// TracerouteTestResult represents the result of a traceroute test
type TracerouteTestResult struct {
	Kind            string          `json:"kind"`
	Name            string          `json:"name"`
	Host            string          `json:"host"`
	Interval        string          `json:"interval"`
	TracerouteStats TracerouteStats `json:"tracerouteStats"`
	Targeted        bool            `json:"targeted"`
	TargetPolicies  []interface{}   `json:"target_policies"`
}

const maxRetries = 3

// fetchTestDetails fetches and processes the details of a single traceroute test
func fetchTestDetails(ctx context.Context, accountID string, testID string, wg *sync.WaitGroup) {
	defer wg.Done()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/dex/traceroute-tests/%s", accountID, testID)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			log.Printf("Error creating request for test %s: %v", testID, err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+config.ApiKey)
		req.Header.Set("Content-Type", "application/json")

		q := req.URL.Query()
		q.Add("timeEnd", time.Now().Format(time.RFC3339))
		q.Add("timeStart", time.Now().Add(-time.Hour).Format(time.RFC3339))
		q.Add("interval", "minute")
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error fetching traceroute test %s: %v", testID, err)
			appmetrics.IncApiErrorsCounter()
			appmetrics.SetUpMetric(0)
			time.Sleep(time.Second * time.Duration(attempt*attempt)) // Exponential backoff
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusGatewayTimeout {
			log.Printf("Service unavailable for test %s: %s", testID, resp.Status)
			time.Sleep(time.Second * time.Duration(attempt*attempt)) // Exponential backoff
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error fetching traceroute test %s: %s", testID, resp.Status)
			appmetrics.IncApiErrorsCounter()
			appmetrics.SetUpMetric(0)
			return
		}

		var response TracerouteTestResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Printf("Error decoding response for test %s: %v", testID, err)
			appmetrics.IncApiErrorsCounter()
			appmetrics.SetUpMetric(0)
			return
		}

		if !response.Success {
			log.Printf("Error in response for test %s: %v", testID, response.Messages)
			appmetrics.IncApiErrorsCounter()
			return
		}

		// Skip HTTP tests
		if response.Result.Kind != "traceroute" {
			return
		}

		stats := response.Result.TracerouteStats

		// Find the most recent slot values
		var latestRTT, latestHops, latestPacketLoss, latestAvailability struct {
			Timestamp string  `json:"timestamp"`
			Value     float64 `json:"value"`
		}

		for _, slot := range stats.RoundTripTimeMs.Slots {
			if slot.Timestamp > latestRTT.Timestamp {
				latestRTT = slot
			}
		}

		for _, slot := range stats.HopsCount.Slots {
			if slot.Timestamp > latestHops.Timestamp {
				latestHops = slot
			}
		}

		for _, slot := range stats.PacketLossPct.Slots {
			if slot.Timestamp > latestPacketLoss.Timestamp {
				latestPacketLoss = slot
			}
		}

		for _, slot := range stats.AvailabilityPct.Slots {
			if slot.Timestamp > latestAvailability.Timestamp {
				latestAvailability = slot
			}
		}

		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_traceroute_rtt{test_id="%s", test_name="%s", host="%s"}`, testID, response.Result.Name, response.Result.Host), func() float64 { return float64(latestRTT.Value) })
		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_traceroute_hops{test_id="%s", test_name="%s", host="%s"}`, testID, response.Result.Name, response.Result.Host), func() float64 { return float64(latestHops.Value) })
		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_traceroute_packet_loss{test_id="%s", test_name="%s", host="%s"}`, testID, response.Result.Name, response.Result.Host), func() float64 { return float64(latestPacketLoss.Value) })
		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_traceroute_availability{test_id="%s", test_name="%s", host="%s"}`, testID, response.Result.Name, response.Result.Host), func() float64 { return float64(latestAvailability.Value) })

		break
	}
}

// CollectTracerouteMetrics fetches detailed metrics for each traceroute test
func CollectTracerouteMetrics(ctx context.Context, accountID string, testIDs []string) {
	var wg sync.WaitGroup
	wg.Add(len(testIDs))

	for _, testID := range testIDs {
		go fetchTestDetails(ctx, accountID, testID, &wg)
	}

	wg.Wait()
}
