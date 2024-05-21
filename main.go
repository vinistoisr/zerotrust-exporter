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
	// Load environment variables if not set by flags
	apiKey = os.Getenv("API_KEY")
	accountID = os.Getenv("ACCOUNT_ID")
	debug = os.Getenv("DEBUG") == "true"
	enableDevices = os.Getenv("DEVICES") == "true"
	enableUsers = os.Getenv("USERS") == "true"
	enableTunnels = os.Getenv("TUNNELS") == "true"
	listenAddr = os.Getenv("INTERFACE")
	port = 9184 // Default port
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		fmt.Sscanf(portEnv, "%d", &port)
	}

	// Define command-line flags (override env variables if set)
	flag.StringVar(&apiKey, "apikey", apiKey, "Cloudflare API key (required)")
	flag.StringVar(&accountID, "accountid", accountID, "Cloudflare account ID (required)")
	flag.BoolVar(&debug, "debug", debug, "Enable debug mode")
	flag.BoolVar(&enableDevices, "devices", enableDevices, "Enable devices metrics")
	flag.BoolVar(&enableUsers, "users", enableUsers, "Enable users metrics")
	flag.BoolVar(&enableTunnels, "tunnels", enableTunnels, "Enable tunnels metrics")
	flag.StringVar(&listenAddr, "interface", listenAddr, "Listening interface (default: any)")
	flag.IntVar(&port, "port", port, "Listening port (default: 9184)")
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
		log.Printf("API Key: %s", apiKey)
		log.Printf("Account ID: %s", accountID)
		log.Printf("Debug: %v", debug)
		log.Printf("Devices: %v", enableDevices)
		log.Printf("Users: %v", enableUsers)
		log.Printf("Tunnels: %v", enableTunnels)
		log.Printf("Interface: %s", listenAddr)
		log.Printf("Port: %d", port)
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
