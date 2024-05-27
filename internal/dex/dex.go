package dex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/appmetrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/config"
)

// Define the structs for the dex tests
// Define the structs for the dex tests

type History struct {
	AvgMs      int        `json:"avgMs"`
	DeltaPct   float64    `json:"deltaPct"`
	TimePeriod TimePeriod `json:"timePeriod"`
}

type OverTimeValue struct {
	Timestamp string `json:"timestamp"`
	AvgMs     int    `json:"avgMs"`
}

type OverTime struct {
	Values     []OverTimeValue `json:"values"`
	TimePeriod TimePeriod      `json:"timePeriod"`
}

type TimePeriod struct {
	Value int    `json:"value"`
	Units string `json:"units"`
}

type RoundTripTime struct {
	AvgMs    int       `json:"avgMs"`
	History  []History `json:"history"`
	OverTime OverTime  `json:"overTime"`
}

type TracerouteResults struct {
	RoundTripTime RoundTripTime `json:"roundTripTime"`
}

type HTTPResults struct {
	ResourceFetchTime struct {
		AvgMs    int       `json:"avgMs"`
		History  []History `json:"history"`
		OverTime OverTime  `json:"overTime"`
	} `json:"resourceFetchTime"`
}

type DexTests struct {
	TestID            string             `json:"id"`
	TestName          string             `json:"name"`
	Kind              string             `json:"kind"`
	Enabled           bool               `json:"enabled"`
	Description       string             `json:"description"`
	Host              string             `json:"host"`
	TracerouteResults *TracerouteResults `json:"tracerouteResults,omitempty"`
	HTTPResults       *HTTPResults       `json:"httpResults,omitempty"`
}

type ResultInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

type ApiResponse struct {
	Result struct {
		Tests []map[string]interface{} `json:"tests"`
	} `json:"result"`
	Success    bool          `json:"success"`
	Errors     []interface{} `json:"errors"`
	Messages   []interface{} `json:"messages"`
	ResultInfo ResultInfo    `json:"result_info"`
}

// createRequest creates a new http request with the given url, page and perPage
func createRequest(ctx context.Context, url string, page int, perPage int) (*http.Request, error) {
	log.Printf("Creating request for %s", url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("per_page", fmt.Sprintf("%d", perPage))
	q.Add("page", fmt.Sprintf("%d", page))
	q.Add("timeEnd", time.Now().Format(time.RFC3339))
	q.Add("timeStart", time.Now().Add(-time.Hour).Format(time.RFC3339))
	req.URL.RawQuery = q.Encode()

	return req, nil
}

// CollectDexTests fetches all the tests from the dex API
func CollectDexTests(ctx context.Context, accountID string) (map[string]DexTests, error) {
	log.Printf("Fetching dex tests for account %s", accountID)
	startTime := time.Now()
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/dex/tests", accountID)
	page := 1
	perPage := 50

	tests := make(map[string]DexTests)

	for {
		log.Printf("Fetching page %d of dex tests", page)
		req, err := createRequest(ctx, url, page, perPage)
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error fetching dex tests: %v", err)
			appmetrics.IncApiErrorsCounter()
			appmetrics.SetUpMetric(0)
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error fetching dex tests: %s", resp.Status)
			appmetrics.IncApiErrorsCounter()
			appmetrics.SetUpMetric(0)
			return nil, fmt.Errorf("error fetching dex tests: %s", resp.Status)
		}

		var response ApiResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			log.Printf("Error decoding response: %v", err)
			appmetrics.IncApiErrorsCounter()
			appmetrics.SetUpMetric(0)
			return nil, err
		}

		if !response.Success {
			log.Printf("Error fetching dex tests: %v", response.Messages)
			appmetrics.IncApiErrorsCounter()
			return nil, fmt.Errorf("failed to fetch dex tests: %v", response.Messages)
		}

		for _, test := range response.Result.Tests {
			data, err := json.Marshal(test)
			if err != nil {
				log.Printf("Error marshalling test: %v", err)
				continue
			}

			var dexTest DexTests
			if err := json.Unmarshal(data, &dexTest); err != nil {
				log.Printf("Error unmarshalling test: %v", err)
				continue
			}

			tests[dexTest.TestID] = dexTest
		}

		if page >= response.ResultInfo.TotalPages {
			break
		}
		page++
	}

	if config.Debug {
		log.Printf("Fetched %d dex tests in %v", len(tests), time.Since(startTime))
	}

	for _, test := range tests {
		switch test.Kind {
		case "traceroute":
			if test.TracerouteResults != nil {
				for _, h := range test.TracerouteResults.RoundTripTime.History {
					if h.TimePeriod.Value == 1 && h.TimePeriod.Units == "hours" {
						metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_dex_test_1h_avg_ms{test_id="%s", test_name="%s", description="%s", host="%s", kind="%s"}`, test.TestID, test.TestName, test.Description, test.Host, test.Kind), func() float64 { return float64(h.AvgMs) })
					}
				}
			}
		case "http":
			if test.HTTPResults != nil {
				for _, h := range test.HTTPResults.ResourceFetchTime.History {
					if h.TimePeriod.Value == 1 && h.TimePeriod.Units == "hours" {
						metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_dex_test_1h_avg_ms{test_id="%s", test_name="%s", description="%s", host="%s", kind="%s"}`, test.TestID, test.TestName, test.Description, test.Host, test.Kind), func() float64 { return float64(h.AvgMs) })
					}
				}
			}
		}
	}

	return tests, nil
}

// CollectDexMetrics collects metrics for dex
func CollectDexMetrics(ctx context.Context, accountID string) {
	// Collect dex tests
	tests, err := CollectDexTests(ctx, accountID)
	if err != nil {
		log.Printf("Error collecting dex metrics: %v", err)
		appmetrics.IncApiErrorsCounter()
		appmetrics.SetUpMetric(0)
		return
	}

	if config.Debug {
		log.Printf("Fetched %d dex tests", len(tests))
	}
	// Collect test IDs
	testIDs := make([]string, 0, len(tests))
	for testID := range tests {
		testIDs = append(testIDs, testID)
	}
	// Collect traceroute metrics
	CollectTracerouteMetrics(ctx, accountID, testIDs)
}
