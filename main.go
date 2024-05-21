package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cloudflare/cloudflare-go"
)

// Command-line flags
var (
	apiKey        string
	accountID     string
	debug         bool
	enableDevices bool
	enableUsers   bool
	enableTunnels bool
	listenAddr    string
	port          int
	client        *cloudflare.API
)

func init() {
	// Define command-line flags
	flag.StringVar(&apiKey, "apikey", "", "Cloudflare API key (required)")
	flag.StringVar(&accountID, "accountid", "", "Cloudflare account ID (required)")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.BoolVar(&enableDevices, "devices", false, "Enable devices metrics")
	flag.BoolVar(&enableUsers, "users", false, "Enable users metrics")
	flag.BoolVar(&enableTunnels, "tunnels", false, "Enable tunnels metrics")
	flag.StringVar(&listenAddr, "interface", "", "Listening interface (default: any)")
	flag.IntVar(&port, "port", 9184, "Listening port (default: 9184)")
	flag.Parse()

	// Ensure required flags are provided
	if apiKey == "" || accountID == "" {
		fmt.Println("Both apikey and accountid are required")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize Cloudflare client
	var err error
	client, err = cloudflare.NewWithAPIToken(apiKey)
	if err != nil {
		log.Fatalf("Failed to create Cloudflare client: %v", err)
	}
}

func main() {
	addr := fmt.Sprintf("%s:%d", listenAddr, port)
	if debug {
		// Print debug information on startup
		log.Printf("Starting server on %s with debug mode enabled", addr)
		log.Printf("Devices metrics enabled: %v", enableDevices)
		log.Printf("Users metrics enabled: %v", enableUsers)
		log.Printf("Tunnels metrics enabled: %v", enableTunnels)
		log.Printf("API Key: %s", os.Getenv("API_KEY"))
		log.Printf("Account ID: %s", os.Getenv("ACCOUNT_ID"))
		log.Printf("Debug: %s", os.Getenv("DEBUG"))
		log.Printf("Devices: %s", os.Getenv("DEVICES"))
		log.Printf("Users: %s", os.Getenv("USERS"))
		log.Printf("Tunnels: %s", os.Getenv("TUNNELS"))
		log.Printf("Interface: %s", os.Getenv("INTERFACE"))
		log.Printf("Port: %s", os.Getenv("PORT"))
	} else {
		// Print normal startup message
		log.Printf("Starting server on %s", addr)
	}

	// Register metrics handler
	http.HandleFunc("/metrics", metricsHandler)
	// Start HTTP server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
