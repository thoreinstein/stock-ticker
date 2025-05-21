package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

// MockHTTPClient is a mock implementation of HTTPClient for testing
type MockHTTPClient struct {
	DoFunc func(url string) (*http.Response, error)
}

// Get is the mock implementation of HTTPClient.Get
func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.DoFunc(url)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expected    *Config
		expectError bool
	}{
		{
			name: "All environment variables set correctly",
			envVars: map[string]string{
				"SYMBOL": "AAPL",
				"NDAYS":  "5",
				"APIKEY": "test-api-key",
			},
			expected: &Config{
				Symbol: "AAPL",
				NDays:  5,
				APIKey: "test-api-key",
			},
			expectError: false,
		},
		{
			name: "Missing SYMBOL",
			envVars: map[string]string{
				"SYMBOL": "",
				"NDAYS":  "5",
				"APIKEY": "test-api-key",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Missing NDAYS",
			envVars: map[string]string{
				"SYMBOL": "AAPL",
				"NDAYS":  "",
				"APIKEY": "test-api-key",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid NDAYS",
			envVars: map[string]string{
				"SYMBOL": "AAPL",
				"NDAYS":  "invalid",
				"APIKEY": "test-api-key",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Missing APIKEY",
			envVars: map[string]string{
				"SYMBOL": "AAPL",
				"NDAYS":  "5",
				"APIKEY": "",
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment variables
			origSymbol := os.Getenv("SYMBOL")
			origNDays := os.Getenv("NDAYS")
			origAPIKey := os.Getenv("APIKEY")

			// Restore original environment variables after test
			defer func() {
				os.Setenv("SYMBOL", origSymbol)
				os.Setenv("NDAYS", origNDays)
				os.Setenv("APIKEY", origAPIKey)
			}()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			config, err := loadConfig()

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check config
			if !tt.expectError {
				if config == nil {
					t.Error("Expected config, got nil")
				} else if !reflect.DeepEqual(config, tt.expected) {
					t.Errorf("Expected config %+v, got %+v", tt.expected, config)
				}
			}
		})
	}
}

func TestCreateHandler(t *testing.T) {
	config := &Config{
		Symbol: "AAPL",
		NDays:  5,
		APIKey: "test-api-key",
	}

	tests := []struct {
		name           string
		method         string
		mockClient     *MockHTTPClient
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "GET request returns stock data",
			method: http.MethodGet,
			mockClient: &MockHTTPClient{
				DoFunc: func(url string) (*http.Response, error) {
					mockResponse := AlphaVantageResponse{
						MetaData: map[string]interface{}{
							"2. Symbol": "AAPL",
						},
						TimeSeries: map[string]map[string]interface{}{
							"2025-01-15": {
								"1. open":   "234.50",
								"2. high":   "236.80",
								"3. low":    "233.20",
								"4. close":  "235.60",
								"5. volume": "45000000",
							},
						},
					}
					jsonData, _ := json.Marshal(mockResponse)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBuffer(jsonData)),
					}, nil
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"symbol":"AAPL"`,
		},
		{
			name:           "POST request returns method not allowed",
			method:         http.MethodPost,
			mockClient:     &MockHTTPClient{},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "PUT request returns method not allowed",
			method:         http.MethodPut,
			mockClient:     &MockHTTPClient{},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:           "DELETE request returns method not allowed",
			method:         http.MethodDelete,
			mockClient:     &MockHTTPClient{},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed",
		},
		{
			name:   "Error fetching stock data",
			method: http.MethodGet,
			mockClient: &MockHTTPClient{
				DoFunc: func(url string) (*http.Response, error) {
					return nil, fmt.Errorf("network error")
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Error fetching stock data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create response recorder
			recorder := httptest.NewRecorder()

			// Create handler and serve request
			handler := createHandler(config, tt.mockClient)
			handler.ServeHTTP(recorder, req)

			// Check status code
			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, recorder.Code)
			}

			// Check response body
			if !strings.Contains(recorder.Body.String(), tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'", tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func TestFetchStockData(t *testing.T) {
	tests := []struct {
		name           string
		symbol         string
		nDays          int
		mockResponse   interface{}
		statusCode     int
		expectedAvg    float64
		expectDataLen  int
		expectedErrMsg string
	}{
		{
			name:          "Valid response with single day",
			symbol:        "AAPL",
			nDays:         1,
			statusCode:    http.StatusOK,
			expectDataLen: 1,
			mockResponse: AlphaVantageResponse{
				MetaData: map[string]interface{}{
					"2. Symbol": "AAPL",
				},
				TimeSeries: map[string]map[string]interface{}{
					"2025-01-15": {
						"1. open":   "234.50",
						"2. high":   "236.80",
						"3. low":    "233.20",
						"4. close":  "235.60",
						"5. volume": "45000000",
					},
				},
			},
			expectedAvg:    235.60,
			expectedErrMsg: "",
		},
		{
			name:          "Valid response with multiple days",
			symbol:        "AAPL",
			nDays:         2,
			statusCode:    http.StatusOK,
			expectDataLen: 2,
			mockResponse: AlphaVantageResponse{
				MetaData: map[string]interface{}{
					"2. Symbol": "AAPL",
				},
				TimeSeries: map[string]map[string]interface{}{
					"2025-01-15": {
						"1. open":   "234.50",
						"2. high":   "236.80",
						"3. low":    "233.20",
						"4. close":  "235.60",
						"5. volume": "45000000",
					},
					"2025-01-14": {
						"1. open":   "232.50",
						"2. high":   "234.80",
						"3. low":    "231.20",
						"4. close":  "233.60",
						"5. volume": "43000000",
					},
				},
			},
			expectedAvg:    234.60, // (235.60 + 233.60) / 2
			expectedErrMsg: "",
		},
		{
			name:          "Request more days than available",
			symbol:        "AAPL",
			nDays:         5,
			statusCode:    http.StatusOK,
			expectDataLen: 1,
			mockResponse: AlphaVantageResponse{
				MetaData: map[string]interface{}{
					"2. Symbol": "AAPL",
				},
				TimeSeries: map[string]map[string]interface{}{
					"2025-01-15": {
						"1. open":   "234.50",
						"2. high":   "236.80",
						"3. low":    "233.20",
						"4. close":  "235.60",
						"5. volume": "45000000",
					},
				},
			},
			expectedAvg:    235.60,
			expectedErrMsg: "",
		},
		{
			name:          "No time series data",
			symbol:        "INVALID",
			nDays:         1,
			statusCode:    http.StatusOK,
			expectDataLen: 0,
			mockResponse: AlphaVantageResponse{
				MetaData:   map[string]interface{}{},
				TimeSeries: nil,
			},
			expectedAvg:    0,
			expectedErrMsg: "no time series data returned",
		},
		{
			name:           "HTTP error",
			symbol:         "AAPL",
			nDays:          1,
			statusCode:     http.StatusInternalServerError,
			expectDataLen:  0,
			mockResponse:   nil,
			expectedAvg:    0,
			expectedErrMsg: "mock HTTP error",
		},
		{
			name:           "Invalid response format",
			symbol:         "AAPL",
			nDays:          1,
			statusCode:     http.StatusOK,
			expectDataLen:  0,
			mockResponse:   "{invalid json}",
			expectedAvg:    0,
			expectedErrMsg: "error decoding",
		},
		{
			name:          "Invalid close price format",
			symbol:        "AAPL",
			nDays:         1,
			statusCode:    http.StatusOK,
			expectDataLen: 0,
			mockResponse: AlphaVantageResponse{
				MetaData: map[string]interface{}{
					"2. Symbol": "AAPL",
				},
				TimeSeries: map[string]map[string]interface{}{
					"2025-01-15": {
						"1. open":   "234.50",
						"2. high":   "236.80",
						"3. low":    "233.20",
						"4. close":  "invalid",
						"5. volume": "45000000",
					},
				},
			},
			expectedAvg:    0, // No data so average is 0
			expectedErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			client := &MockHTTPClient{
				DoFunc: func(url string) (*http.Response, error) {
					if tt.statusCode == http.StatusInternalServerError {
						return nil, fmt.Errorf("mock HTTP error")
					}

					var responseBody []byte
					if mockString, ok := tt.mockResponse.(string); ok {
						responseBody = []byte(mockString)
					} else {
						responseBody, _ = json.Marshal(tt.mockResponse)
					}

					return &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
					}, nil
				},
			}

			// Call function under test
			data, avgClose, err := fetchStockData(tt.symbol, tt.nDays, "dummy-api-key", client)

			// Check error
			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.expectedErrMsg)
				} else if !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErrMsg, err.Error())
				}
				return
			}

			// For valid cases
			if tt.expectedErrMsg == "" && tt.statusCode == http.StatusOK {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}

				// Check data length
				if len(data) != tt.expectDataLen {
					t.Errorf("Expected %d data points, got %d", tt.expectDataLen, len(data))
				}

				// Check average with a small tolerance for floating point rounding
				if tt.expectedAvg != 0 && (avgClose < tt.expectedAvg-0.001 || avgClose > tt.expectedAvg+0.001) {
					t.Errorf("Expected average close %f, got %f", tt.expectedAvg, avgClose)
				}
			}
		})
	}
}

func TestProcessTimeSeries(t *testing.T) {
	tests := []struct {
		name        string
		timeSeries  map[string]map[string]interface{}
		nDays       int
		expectedLen int
		expectedAvg float64
	}{
		{
			name: "Single day",
			timeSeries: map[string]map[string]interface{}{
				"2025-01-15": {
					"1. open":   "234.50",
					"2. high":   "236.80",
					"3. low":    "233.20",
					"4. close":  "235.60",
					"5. volume": "45000000",
				},
			},
			nDays:       1,
			expectedLen: 1,
			expectedAvg: 235.60,
		},
		{
			name: "Multiple days",
			timeSeries: map[string]map[string]interface{}{
				"2025-01-15": {
					"1. open":   "234.50",
					"2. high":   "236.80",
					"3. low":    "233.20",
					"4. close":  "235.60",
					"5. volume": "45000000",
				},
				"2025-01-14": {
					"1. open":   "232.50",
					"2. high":   "234.80",
					"3. low":    "231.20",
					"4. close":  "233.60",
					"5. volume": "43000000",
				},
			},
			nDays:       2,
			expectedLen: 2,
			expectedAvg: 234.60,
		},
		{
			name: "More days requested than available",
			timeSeries: map[string]map[string]interface{}{
				"2025-01-15": {
					"1. open":   "234.50",
					"2. high":   "236.80",
					"3. low":    "233.20",
					"4. close":  "235.60",
					"5. volume": "45000000",
				},
			},
			nDays:       5,
			expectedLen: 1,
			expectedAvg: 235.60,
		},
		{
			name:        "Empty time series",
			timeSeries:  map[string]map[string]interface{}{},
			nDays:       5,
			expectedLen: 0,
			expectedAvg: 0,
		},
		{
			name: "Invalid close price",
			timeSeries: map[string]map[string]interface{}{
				"2025-01-15": {
					"1. open":   "234.50",
					"2. high":   "236.80",
					"3. low":    "233.20",
					"4. close":  "invalid",
					"5. volume": "45000000",
				},
			},
			nDays:       1,
			expectedLen: 0,
			expectedAvg: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call function under test
			data, avgClose := processTimeSeries(tt.timeSeries, tt.nDays)

			// Check data length
			if len(data) != tt.expectedLen {
				t.Errorf("Expected %d data points, got %d", tt.expectedLen, len(data))
			}

			// Check average with a small tolerance for floating point rounding
			if tt.expectedAvg != 0 && (avgClose < tt.expectedAvg-0.001 || avgClose > tt.expectedAvg+0.001) {
				t.Errorf("Expected average close %f, got %f", tt.expectedAvg, avgClose)
			}
		})
	}
}

func TestResponseEncoding(t *testing.T) {
	resp := StockResponse{
		Symbol:       "AAPL",
		Days:         5,
		AverageClose: 235.60,
		Data: []TimeSeriesData{
			{
				Date:       "2025-01-15",
				OpenPrice:  234.50,
				HighPrice:  236.80,
				LowPrice:   233.20,
				ClosePrice: 235.60,
				Volume:     45000000,
			},
		},
	}

	// Encode to JSON
	var buf strings.Builder
	err := json.NewEncoder(&buf).Encode(resp)
	if err != nil {
		t.Fatalf("Failed to encode response: %v", err)
	}

	// Check JSON structure
	jsonStr := buf.String()
	
	expectedFields := []string{
		`"symbol":"AAPL"`,
		`"days":5`,
		`"average_close":235.6`,
		`"data":[`,
		`"date":"2025-01-15"`,
		`"open":234.5`,
		`"high":236.8`,
		`"low":233.2`,
		`"close":235.6`,
		`"volume":45000000`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("Expected JSON to contain '%s', got: %s", field, jsonStr)
		}
	}

	// Decode back
	var decoded StockResponse
	if err := json.NewDecoder(strings.NewReader(jsonStr)).Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// Check fields match
	if decoded.Symbol != resp.Symbol {
		t.Errorf("Expected Symbol %s, got %s", resp.Symbol, decoded.Symbol)
	}
	
	if decoded.Days != resp.Days {
		t.Errorf("Expected Days %d, got %d", resp.Days, decoded.Days)
	}
	
	if decoded.AverageClose != resp.AverageClose {
		t.Errorf("Expected AverageClose %f, got %f", resp.AverageClose, decoded.AverageClose)
	}
	
	if len(decoded.Data) != len(resp.Data) {
		t.Errorf("Expected %d data points, got %d", len(resp.Data), len(decoded.Data))
	} else if len(decoded.Data) > 0 {
		if decoded.Data[0].Date != resp.Data[0].Date {
			t.Errorf("Expected Date %s, got %s", resp.Data[0].Date, decoded.Data[0].Date)
		}
		
		if decoded.Data[0].ClosePrice != resp.Data[0].ClosePrice {
			t.Errorf("Expected ClosePrice %f, got %f", resp.Data[0].ClosePrice, decoded.Data[0].ClosePrice)
		}
	}
}