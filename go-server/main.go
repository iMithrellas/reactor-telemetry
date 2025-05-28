package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type ReactorStats struct {
	Status              string  `json:"status"`
	EnergyStored        float64 `json:"energyStored"`
	EnergyProducedLast  float64 `json:"energyProducedLastTick"`
	FuelTemp            float64 `json:"fuelTemp"`
	CasingTemp          float64 `json:"casingTemp"`
	FuelAmount          int     `json:"fuelAmount"`
	WasteAmount         int     `json:"wasteAmount"`
	FuelConsumedLast    float64 `json:"fuelConsumedLastTick"`
	FuelReactivity      float64 `json:"fuelReactivity"`
	ComputerID          int     `json:"computerID"`
	ComputerLabel       string  `json:"computerLabel"`
	ControlRodInsertion float64 `json:"controlRodInsertion"`
	ReactorType         string  `json:"reactorType"`
	Timestamp           int64   `json:"timestamp"`
}

var (
	mu      sync.RWMutex
	latest  ReactorStats
	history []ReactorStats

	// Initialized in initInflux()
	influxWrite func(ReactorStats)
)

func initInflux() {
	url := os.Getenv("INFLUX_URL")
	token := os.Getenv("INFLUX_TOKEN")
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")

	log.Printf("🔧 Initializing InfluxDB client...")
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)

	log.Printf("Influx config: %s %s %s %s", url, org, bucket, token[:5]+"...")

	// Test connection
	log.Printf("🔍 Testing InfluxDB connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := client.Health(ctx)
	if err != nil {
		log.Printf("⚠️  InfluxDB health check failed: %v", err)
	} else {
		log.Printf("✅ InfluxDB connected: %s", health.Status)
	}

	influxWrite = func(stats ReactorStats) {
		point := influxdb2.NewPointWithMeasurement("reactor").
			AddTag("label", stats.ComputerLabel).
			AddTag("reactor_type", stats.ReactorType).
			AddField("energy_stored", stats.EnergyStored).
			AddField("energy_produced", stats.EnergyProducedLast).
			AddField("fuel_temp", stats.FuelTemp).
			AddField("casing_temp", stats.CasingTemp).
			AddField("fuel_amount", stats.FuelAmount).
			AddField("waste_amount", stats.WasteAmount).
			AddField("fuel_used", stats.FuelConsumedLast).
			AddField("reactivity", stats.FuelReactivity).
			AddField("rod_insertion", stats.ControlRodInsertion).
			AddField("status", stats.Status).
			SetTime(time.Unix(stats.Timestamp, 0))

		// Run in goroutine to prevent blocking
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := writeAPI.WritePoint(ctx, point)
			if err != nil {
				log.Printf("❌ Failed to write to InfluxDB: %v", err)
			} else {
				log.Printf("✅ Wrote stats to InfluxDB")
			}
		}()
	}

	log.Printf("✅ InfluxDB initialization complete")
}

func postStats(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var stats ReactorStats
	if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
		log.Printf("❌ JSON decode failed: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	stats.Timestamp = time.Now().Unix()

	mu.Lock()
	latest = stats
	history = append(history, stats)
	mu.Unlock()

	influxWrite(stats)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func getStats(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(latest)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	log.Printf("🚀 Starting server...")
	initInflux() // ✅ Set up influx writer

	mux := http.NewServeMux()

	mux.HandleFunc("/reactor", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			postStats(w, r)
		case http.MethodGet:
			getStats(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", loggingMiddleware(mux)))
}
