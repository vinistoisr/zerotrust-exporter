package users

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/cloudflare/cloudflare-go"
	"github.com/vinistoisr/zerotrust-exporter/internal/appmetrics"
	"github.com/vinistoisr/zerotrust-exporter/internal/config"
	"github.com/vinistoisr/zerotrust-exporter/internal/devices"
)

// User struct to hold user information (this should match the structure returned by the Cloudflare API)
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	// Add other fields as necessary
}

// fetchAllUsers fetches all users from Cloudflare API
func fetchAllUsers(ctx context.Context) (map[string]*cloudflare.AccessUser, error) {
	rc := &cloudflare.ResourceContainer{Level: cloudflare.AccountRouteLevel, Identifier: config.AccountID}
	startTime := time.Now()
	usersList, _, err := config.Client.ListAccessUsers(ctx, rc, cloudflare.AccessUserParams{})
	if err != nil {
		return nil, err
	}

	users := make(map[string]*cloudflare.AccessUser)
	for _, user := range usersList {
		users[user.ID] = &user
	}

	if config.Debug {
		log.Printf("Fetched %d users in %v", len(users), time.Since(startTime))
	}

	return users, nil
}

// collectUserMetrics collects metrics for users
func CollectUserMetrics(deviceMetrics map[string]devices.DeviceStatus) {
	log.Println("Starting collectUserMetrics...")
	appmetrics.IncApiCallCounter()

	ctx := context.Background()
	// Fetch users from Cloudflare API
	users, err := fetchAllUsers(ctx)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		appmetrics.IncApiErrorsCounter()
		appmetrics.SetUpMetric(0)
		return
	}

	// Logic to update zerotrust_users_up metric for each user
	for _, user := range users {
		for _, device := range deviceMetrics {
			if user.Email == device.PersonEmail {
				log.Printf("Processing user: %s\n", user.Email)

				gatewaySeat := "false"
				if user.GatewaySeat != nil && *user.GatewaySeat {
					gatewaySeat = "true"
				}
				accessSeat := "false"
				if user.AccessSeat != nil && *user.AccessSeat {
					accessSeat = "true"
				}

				metrics.GetOrCreateGauge(fmt.Sprintf(`zerotrust_users_up{gateway_seat="%s", access_seat="%s", user_id="%s", user_email="%s"}`, gatewaySeat, accessSeat, user.ID, user.Email), func() float64 { return 1 })
				break // Exit the inner loop once a match is found
			}
		}
	}
}
