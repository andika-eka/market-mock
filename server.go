package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIResponse is a standard wrapper for our JSON output
type APIResponse struct {
	// Symbol  string   `json:"symbol"`
	// Candles []Candle `json:"data"` // Re-using your Candle struct from engine.go
	Data    string `json:"data"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func handleGetCandles(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Query Parameters
	// Example URL: /api/candles?symbol=BTC-USD&days=10&interval=24h
	query := r.URL.Query()

	symbol := strings.Split(query.Get("symbol"), ",")
	if len(symbol) == 0 {
		http.Error(w, "no symbol found", http.StatusBadRequest)
		return
	}

	// Parse 'days' (How far back to go?)
	daysStr := query.Get("days")
	days := 30 // Default 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	// Parse 'interval' (e.g., "1h", "24h")
	intervalStr := query.Get("interval")
	interval := 24 * time.Hour // Default Daily
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	// 2. Logic (Connecting to your Engine)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Create the Series using your Factory
	// Estimate capacity to avoid resizing (days * (24h / interval))
	approxRows := int(time.Duration(days) * 24 * time.Hour / interval)
	series := NewCandleSeries(symbol, approxRows)

	// Fill it up
	for t := startDate; t.Before(endDate); t = t.Add(interval) {
		c := GetOHLC(symbol, t, interval)
		series.Add(c)
	}

	// 3. Response (JSON)
	w.Header().Set("Content-Type", "application/json")

	// Construct the response object
	resp := APIResponse{
		Symbol:  series.Symbol,
		Candles: series.Candles,
	}

	// Write to stream
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func startServer() {
	// Go 1.22+ Routing
	mux := http.NewServeMux()

	// Register the handler
	mux.HandleFunc("GET /api/candles", handleGetCandles)

	fmt.Println("🚀 Server running on http://localhost:8080")
	fmt.Println("Try: curl 'http://localhost:8080/api/candles?symbol=ETH-USD&days=5&interval=1h'")

	// Start Listening
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Server failed:", err)
	}
}
