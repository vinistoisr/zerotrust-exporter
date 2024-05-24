package tunnels

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/cloudflare/cloudflare-go"
	"github.com/vinistoisr/zerotrust-exporter/internal/collector"
	"github.com/vinistoisr/zerotrust-exporter/internal/config"
)

// collectTunnelMetrics collects metrics for tunnels
func collectTunnelMetrics() {
	collector.ApiCallCounter.Inc()

	ctx := context.Background()
	rc := &cloudflare.ResourceContainer{Level: cloudflare.AccountRouteLevel, Identifier: config.AccountID}
	startTime := time.Now()
	// Fetch tunnels from Cloudflare API
	tunnels, _, err := config.Client.ListTunnels(ctx, rc, cloudflare.TunnelListParams{})
	if err != nil {
		log.Printf("Error fetching tunnels: %v", err)
		collector.ApiErrorsCounter.Inc()
		collector.UpMetric.Set(0)
		return
	}

	if config.Debug {
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
