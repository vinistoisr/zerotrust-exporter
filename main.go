package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/VictoriaMetrics/metrics"
)

var (
	apiKey string
	addr   string
	debug  bool
)

func init() {
	flag.StringVar(&apiKey, "api-key", os.Getenv("CLOUDFLARE_API_KEY"), "Cloudflare API key")
	flag.StringVar(&addr, "addr", ":9184", "Address to listen on")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
	flag.Parse()

	if apiKey == "" {
		log.Fatal("Cloudflare API key is required")
	}
}

func main() {
	http.HandleFunc("/api", apiHandler)
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics.WritePrometheus(w, true)
	})

	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
