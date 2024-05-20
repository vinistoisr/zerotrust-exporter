package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/cenkalti/backoff/v4"
	"github.com/cloudflare/cloudflare-go"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Missing 'target' parameter", http.StatusBadRequest)
		return
	}

	requestType := r.URL.Query().Get("request_type")
	if requestType == "" {
		requestType = "GET"
	}

	accountID := r.URL.Query().Get("account_id")
	if accountID == "" {
		accountID = getDefaultAccountID()
		if accountID == "" {
			http.Error(w, "No account ID provided and could not determine default account ID", http.StatusBadRequest)
			return
		}
	}

	api, err := cloudflare.NewWithAPIToken(apiKey)
	if err != nil {
		log.Printf("Error creating Cloudflare API client: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Fetch data from the Cloudflare API with retries
	var data interface{}
	operation := func() error {
		return fetchCloudflareData(api, target, requestType, accountID, &data)
	}

	err = backoff.Retry(operation, backoff.NewExponentialBackOff())
	if err != nil {
		log.Printf("Error fetching data from Cloudflare API: %v", err)
		http.Error(w, "Failed to fetch data from Cloudflare API", http.StatusInternalServerError)
		return
	}

	// Generate Prometheus metrics
	generateMetrics(data)

	// Log in debug mode
	if debug {
		log.Printf("Successfully scraped endpoint: %s with request type: %s for account: %s", target, requestType, accountID)
	}

	fmt.Fprintf(w, "Metrics generated successfully")
}

func getDefaultAccountID() string {
	api, err := cloudflare.NewWithAPIToken(apiKey)
	if err != nil {
		log.Printf("Error creating Cloudflare API client: %v", err)
		return ""
	}

	// Create context
	ctx := context.Background()

	// Use default params
	params := cloudflare.AccountsListParams{}

	accounts, _, err := api.Accounts(ctx, params)
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
		return ""
	}

	if len(accounts) > 0 {
		return accounts[0].ID
	}

	return ""
}

func fetchCloudflareData(api *cloudflare.API, endpoint, requestType, accountID string, result interface{}) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %v", err)
	}

	req, err := http.NewRequest(requestType, u.String(), nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
