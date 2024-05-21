package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/cloudflare/cloudflare-go"
)

// Prometheus metrics
var (
	upMetric         = metrics.NewGauge("zerotrust_exporter_up", func() float64 { return 1 })
	scrapeDuration   = metrics.NewHistogram("zerotrust_exporter_scrape_duration_seconds")
	apiCallCounter   = metrics.NewCounter("zerotrust_exporter_api_calls_total")
	apiErrorsCounter = metrics.NewCounter("zerotrust_exporter_api_errors_total")
)

// metricsHandler handles the /metrics endpoint
func metricsHandler(w http.ResponseWriter, req *http.Request) {
	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(3)

	// Collect device metrics
	go func() {
		defer wg.Done()
		if enableDevices {
			collectDeviceMetrics()
		}
	}()
	// Collect user metrics
	go func() {
		defer wg.Done()
		if enableUsers {
			collectUserMetrics()
		}
	}()
	// Collect tunnel metrics
	go func() {
		defer wg.Done()
		if enableTunnels {
			collectTunnelMetrics()
		}
	}()

	// Wait for all metrics collection to complete
	wg.Wait()

	// Update scrape duration metric
	scrapeDuration.UpdateDuration(startTime)
	// Write metrics to the response
	metrics.WritePrometheus(w, true)

	// Print debug information if enabled
	if debug {
		log.Printf("Scrape completed in %v", time.Since(startTime))
		log.Printf("API calls made: %d", apiCallCounter.Get())
		log.Printf("API errors encountered: %d", apiErrorsCounter.Get())
	}
}

// collectDeviceMetrics collects metrics for devices
func collectDeviceMetrics() {
	apiCallCounter.Inc()

	ctx := context.Background()
	startTime := time.Now()
	// Fetch devices from Cloudflare API
	devices, err := client.ListTeamsDevices(ctx, accountID)
	if err != nil {
		log.Printf("Error fetching devices: %v", err)
		apiErrorsCounter.Inc()
		upMetric.Set(0)
		return
	}

	if debug {
		log.Printf("Fetched %d devices in %v", len(devices), time.Since(startTime))
	}

	// Fetch all users to get their email addresses
	users, err := fetchAllUsers(ctx)
	if err != nil {
		log.Printf("Error fetching user details: %v", err)
		apiErrorsCounter.Inc()
		upMetric.Set(0)
		return
	}

	// Collect metrics for each device
	for _, device := range devices {
		lastSeen, err := time.Parse(time.RFC3339, device.LastSeen)
		if err != nil {
			log.Printf("Error parsing last seen time: %v", err)
			continue
		}
		deviceUp := 0
		if time.Since(lastSeen) <= 5*time.Minute {
			deviceUp = 1
		}

		userEmail := ""
		if device.User.ID != "" {
			if user, ok := users[device.User.ID]; ok {
				userEmail = user.Email
			}
		}

		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_devices_up{device_type="%s", id="%s", ip="%s", user_id="%s", user_email="%s", name="%s"}`, device.DeviceType, device.ID, device.IP, device.User.ID, userEmail, device.Name), func() float64 { return float64(deviceUp) })
	}
}

// fetchAllUsers fetches all users from Cloudflare API
func fetchAllUsers(ctx context.Context) (map[string]*cloudflare.AccessUser, error) {
	rc := &cloudflare.ResourceContainer{Level: cloudflare.AccountRouteLevel, Identifier: accountID}
	startTime := time.Now()
	usersList, _, err := client.ListAccessUsers(ctx, rc, cloudflare.AccessUserParams{})
	if err != nil {
		return nil, err
	}

	users := make(map[string]*cloudflare.AccessUser)
	for _, user := range usersList {
		users[user.ID] = &user
	}

	if debug {
		log.Printf("Fetched %d users in %v", len(users), time.Since(startTime))
	}

	return users, nil
}

// collectUserMetrics collects metrics for users
func collectUserMetrics() {
	apiCallCounter.Inc()

	ctx := context.Background()
	rc := &cloudflare.ResourceContainer{Level: cloudflare.AccountRouteLevel, Identifier: accountID}
	startTime := time.Now()
	// Fetch users from Cloudflare API
	users, _, err := client.ListAccessUsers(ctx, rc, cloudflare.AccessUserParams{})
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		apiErrorsCounter.Inc()
		upMetric.Set(0)
		return
	}

	if debug {
		log.Printf("Fetched %d users in %v", len(users), time.Since(startTime))
	}

	// Collect metrics for each user
	for _, user := range users {
		gatewaySeat := "false"
		if user.GatewaySeat != nil && *user.GatewaySeat {
			gatewaySeat = "true"
		}
		accessSeat := float64(0)
		if user.AccessSeat != nil && *user.AccessSeat {
			accessSeat = 1
		}
		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_users_access_seat{email="%s", id="%s", gateway_seat="%s"}`, user.Email, user.ID, gatewaySeat), func() float64 { return accessSeat })
	}
}

// collectTunnelMetrics collects metrics for tunnels
func collectTunnelMetrics() {
	apiCallCounter.Inc()

	ctx := context.Background()
	rc := &cloudflare.ResourceContainer{Level: cloudflare.AccountRouteLevel, Identifier: accountID}
	startTime := time.Now()
	// Fetch tunnels from Cloudflare API
	tunnels, _, err := client.ListTunnels(ctx, rc, cloudflare.TunnelListParams{})
	if err != nil {
		log.Printf("Error fetching tunnels: %v", err)
		apiErrorsCounter.Inc()
		upMetric.Set(0)
		return
	}

	if debug {
		log.Printf("Fetched %d tunnels in %v", len(tunnels), time.Since(startTime))
	}

	// Collect metrics for each tunnel
	for _, tunnel := range tunnels {
		status := 0
		if tunnel.Status == "healthy" {
			status = 1
		}
		metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_tunnels_status{id="%s", name="%s"}`, tunnel.ID, tunnel.Name), func() float64 { return float64(status) })
	}
}
