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
	EnergyStored        int     `json:"energyStored"`
	EnergyProducedLast  float64 `json:"energyProducedLastTick"`
	FuelTemp            int     `json:"fuelTemp"`
	CasingTemp          int     `json:"casingTemp"`
	FuelAmount          int     `json:"fuelAmount"`
	WasteAmount         int     `json:"wasteAmount"`
	FuelConsumedLast    float64 `json:"fuelConsumedLastTick"`
	FuelReactivity      int     `json:"fuelReactivity"`
	ComputerID          int     `json:"computerID"`
	ComputerLabel       string  `json:"computerLabel"`
	ControlRodInsertion int     `json:"controlRodInsertion"`
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

	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)

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

		err := writeAPI.WritePoint(context.Background(), point)
		if err != nil {
			log.Printf("❌ Failed to write to InfluxDB: %v", err)
		} else {
			log.Println("✅ Wrote stats to InfluxDB")
		}
	}
}

func postStats(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	log.Printf("Received raw: %s", string(bodyBytes))
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var stats ReactorStats
	if err := json.NewDecoder(r.Body).Decode(&stats); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	stats.Timestamp = time.Now().Unix()

	mu.Lock()
	latest = stats
	history = append(history, stats)
	mu.Unlock()

	influxWrite(stats) // 🔥 Push to InfluxDB

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

func getHistory(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
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

	mux.HandleFunc("/history", getHistory)
	mux.HandleFunc("/ping", ping)

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", loggingMiddleware(mux)))
}
