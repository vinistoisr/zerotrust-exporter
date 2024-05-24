package devices

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/cloudflare/cloudflare-go"
	"github.com/vinistoisr/zerotrust-exporter/internal/appmetrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/config"
)

type DeviceStatus struct {
	Colo        string `json:"colo"`
	Mode        string `json:"mode"`
	Status      string `json:"status"`
	Platform    string `json:"platform"`
	Version     string `json:"version"`
	Timestamp   string `json:"timestamp"`
	DeviceName  string `json:"device_name"`
	DeviceID    string `json:"device_id"`
	PersonEmail string `json:"person_email"`
}

func fetchDeviceStatus(ctx context.Context, client *cloudflare.API, accountID string) (map[string]DeviceStatus, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/dex/fleet-status/devices", accountID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+config.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("per_page", "50")
	q.Add("page", "1")
	q.Add("time_end", time.Unix(time.Now().Unix(), 0).Format(time.RFC3339))
	q.Add("time_start", time.Unix(time.Now().Add(-time.Minute*3).Unix(), 0).Format(time.RFC3339))
	q.Add("sort_by", "device_id")
	q.Add("status", "connected")
	q.Add("source", "last_seen")

	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	appmetrics.IncApiCallCounter()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("failed to fetch device status: %s, response body: %s", resp.Status, bodyString)
	}

	var response struct {
		Result []DeviceStatus `json:"result"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	deviceStatuses := make(map[string]DeviceStatus)
	for _, deviceStatus := range response.Result {
		deviceStatuses[deviceStatus.DeviceID] = deviceStatus
	}

	return deviceStatuses, nil
}

func CollectDeviceMetrics() map[string]DeviceStatus {
	log.Println("Collecting device metrics...")
	appmetrics.IncApiCallCounter()
	ctx := context.Background()
	startTime := time.Now()

	deviceStatuses, err := fetchDeviceStatus(ctx, config.Client, config.AccountID)
	if err != nil {
		log.Printf("Error fetching device status: %v", err)
		appmetrics.IncApiErrorsCounter()
		appmetrics.SetUpMetric(0)
		return nil
	}

	if config.Debug {
		log.Printf("Fetched %d devices in %v", len(deviceStatuses), time.Since(startTime))
	}

	filteredDevices := make(map[string]DeviceStatus)
	for _, status := range deviceStatuses {
		if status.Status == "connected" {
			filteredDevices[status.DeviceID] = status
		}
	}

	for deviceID, status := range filteredDevices {
		deviceUp := 1
		metricName := fmt.Sprintf(`zerotrust_devices_status{device_id="%s", device_name="%s", user_email="%s", colo="%s", mode="%s", platform="%s", version="%s"}`, deviceID, status.DeviceName, status.PersonEmail, status.Colo, status.Mode, status.Platform, status.Version)
		gauge := metrics.GetOrCreateGauge(metricName, func() float64 { return float64(deviceUp) })
		gauge.Set(float64(deviceUp))
	}

	log.Println("Device metrics collection completed.")
	return filteredDevices
}
