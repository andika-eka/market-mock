package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type APIResponse struct {
	Data    any    `json:"data"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func HandleGetCandles(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	symbols := strings.Split(query.Get("symbol"), ",")
	daysStr := query.Get("days")

	if len(symbols) == 0 {
		http.Error(w, "no symbol found", http.StatusBadRequest)
		return
	}

	days := 30 //default
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil {
			days = d
		}
	}

	intervalStr := query.Get("interval")
	interval := 24 * time.Hour // Default
	if intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			interval = d
		}
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	symseries := make([]*CandleSeries, 0, len(symbols))
	for _, symbol := range symbols {
		series := GetDataSeries(symbol, startDate, endDate, interval)
		symseries = append(symseries, series)
	}

	w.Header().Set("Content-Type", "application/json")

	// Construct the response object
	resp := APIResponse{
		Data:    symseries,
		Success: true,
		Message: "Success",
	}

	// Write to stream
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func StartServer() {
	// Go 1.22+ Routing
	mux := http.NewServeMux()

	// Register the handler
	mux.HandleFunc("GET /api/candles", HandleGetCandles)

	fmt.Println("Server running on http://localhost:8080")
	fmt.Println("Try: curl 'http://localhost:8080/api/candles?symbol=ETH-USD&days=5&interval=1h'")

	// Start Listening
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Println("Server failed:", err)
	}
}
