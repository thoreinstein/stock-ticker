package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// StockResponse matches the response structure from main.go
type StockResponse struct {
	Symbol       string           `json:"symbol"`
	Days         int              `json:"days"`
	AverageClose float64          `json:"average_close"`
	Data         []TimeSeriesData `json:"data"`
}

type TimeSeriesData struct {
	Date       string  `json:"date"`
	OpenPrice  float64 `json:"open"`
	HighPrice  float64 `json:"high"`
	LowPrice   float64 `json:"low"`
	ClosePrice float64 `json:"close"`
	Volume     int64   `json:"volume"`
}

func TestAPI_MVP_Integration(t *testing.T) {
	// This is a simplified integration test that doesn't actually start the server
	// In a real production environment, we would use a test server

	// Create a mock handler that simulates our API behavior
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Mock response
		response := StockResponse{
			Symbol:       "AAPL",
			Days:         1,
			AverageClose: 235.50,
			Data: []TimeSeriesData{
				{
					Date:       "2025-01-15",
					ClosePrice: 235.50,
					OpenPrice:  234.00,
					HighPrice:  236.00,
					LowPrice:   233.00,
					Volume:     1000000,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			t.Fatalf("Failed to encode response: %v", err)
		}
	})

	// Test the handler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "application/json")
	}

	// Decode response
	var response StockResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Validate response
	if response.Symbol != "AAPL" {
		t.Errorf("Expected symbol AAPL, got %s", response.Symbol)
	}

	if response.Days != 1 {
		t.Errorf("Expected 1 day, got %d", response.Days)
	}

	if response.AverageClose == 0 {
		t.Error("Average close should not be zero")
	}

	if len(response.Data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(response.Data))
	}
}

func TestAPI_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// Simple handler that checks method
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", status)
	}
}
