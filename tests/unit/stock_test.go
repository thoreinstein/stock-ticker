package unit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock the Alpha Vantage response structure
type MockAlphaVantageResponse struct {
	MetaData   map[string]interface{}            `json:"Meta Data"`
	TimeSeries map[string]map[string]interface{} `json:"Time Series (Daily)"`
}

func TestAlphaVantageResponseParsing(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := MockAlphaVantageResponse{
			MetaData: map[string]interface{}{
				"1. Information": "Daily Prices",
				"2. Symbol":      "AAPL",
			},
			TimeSeries: map[string]map[string]interface{}{
				"2025-01-15": {
					"1. open":   "235.50",
					"2. high":   "237.80",
					"3. low":    "234.20",
					"4. close":  "236.60",
					"5. volume": "50000000",
				},
			},
		}
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			t.Fatal("Failed to encode response:", err)
		}
	}))
	defer server.Close()

	// Make request to mock server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	// Decode response
	var avResp MockAlphaVantageResponse
	if err := json.NewDecoder(resp.Body).Decode(&avResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Validate parsing
	if avResp.TimeSeries == nil {
		t.Error("Time series should not be nil")
	}

	if len(avResp.TimeSeries) != 1 {
		t.Errorf("Expected 1 time series entry, got %d", len(avResp.TimeSeries))
	}

	// Test close price parsing
	dayData, exists := avResp.TimeSeries["2025-01-15"]
	if !exists {
		t.Error("Expected date 2025-01-15 not found")
	}

	closePrice, ok := dayData["4. close"].(string)
	if !ok {
		t.Error("Close price not found or wrong type")
	}

	if closePrice != "236.60" {
		t.Errorf("Expected close price 236.60, got %s", closePrice)
	}
}

func TestAverageCalculation(t *testing.T) {
	testCases := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "Single value",
			values:   []float64{100.0},
			expected: 100.0,
		},
		{
			name:     "Multiple values",
			values:   []float64{100.0, 200.0, 300.0},
			expected: 200.0,
		},
		{
			name:     "Decimal values",
			values:   []float64{10.5, 20.5},
			expected: 15.5,
		},
		{
			name:     "Empty slice",
			values:   []float64{},
			expected: 0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var total float64
			for _, v := range tc.values {
				total += v
			}

			var avg float64
			if len(tc.values) > 0 {
				avg = total / float64(len(tc.values))
			}

			if avg != tc.expected {
				t.Errorf("Expected average %f, got %f", tc.expected, avg)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	// Test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("invalid json")); err != nil {
			t.Fatal("Failed to write response:", err)
		}
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	var avResp MockAlphaVantageResponse
	err = json.NewDecoder(resp.Body).Decode(&avResp)
	if err == nil {
		t.Error("Expected error parsing invalid JSON, got nil")
	}
}

func TestEmptyTimeSeries(t *testing.T) {
	// Test server that returns empty time series
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := MockAlphaVantageResponse{
			MetaData: map[string]interface{}{
				"1. Information": "Daily Prices",
				"2. Symbol":      "PING",
			},
			TimeSeries: nil,
		}
		if err := json.NewEncoder(w).Encode(mockResponse); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			t.Fatal("Failed to encode response:", err)
		}
	}))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	}()

	var avResp MockAlphaVantageResponse
	if err := json.NewDecoder(resp.Body).Decode(&avResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if avResp.TimeSeries != nil {
		t.Error("Expected nil time series")
	}
}
