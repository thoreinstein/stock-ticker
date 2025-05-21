package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// TimeSeriesData represents a single day of stock data
type TimeSeriesData struct {
	Date       string  `json:"date"`
	OpenPrice  float64 `json:"open"`
	HighPrice  float64 `json:"high"`
	LowPrice   float64 `json:"low"`
	ClosePrice float64 `json:"close"`
	Volume     int64   `json:"volume"`
}

// StockResponse is the API response format
type StockResponse struct {
	Symbol       string           `json:"symbol"`
	Days         int              `json:"days"`
	AverageClose float64          `json:"average_close"`
	Data         []TimeSeriesData `json:"data"`
}

// AlphaVantageResponse is the format returned by the Alpha Vantage API
type AlphaVantageResponse struct {
	MetaData   map[string]interface{}            `json:"Meta Data"`
	TimeSeries map[string]map[string]interface{} `json:"Time Series (Daily)"`
}

// Config holds the application configuration
type Config struct {
	Symbol string
	NDays  int
	APIKey string
}

// HTTPClient interface allows us to mock the http.Client in tests
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// DefaultHTTPClient is the default implementation of HTTPClient
type DefaultHTTPClient struct{}

// Get implements the HTTPClient interface
func (c *DefaultHTTPClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	client := &DefaultHTTPClient{}
	startServer(config, client)
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	symbol := os.Getenv("SYMBOL")
	if symbol == "" {
		return nil, fmt.Errorf("SYMBOL environment variable is required")
	}

	nDaysStr := os.Getenv("NDAYS")
	if nDaysStr == "" {
		return nil, fmt.Errorf("NDAYS environment variable is required")
	}

	nDays, err := strconv.Atoi(nDaysStr)
	if err != nil {
		return nil, fmt.Errorf("Invalid NDAYS value: %v", err)
	}

	apiKey := os.Getenv("APIKEY")
	if apiKey == "" {
		return nil, fmt.Errorf("APIKEY environment variable is required")
	}

	return &Config{
		Symbol: symbol,
		NDays:  nDays,
		APIKey: apiKey,
	}, nil
}

// startServer starts the HTTP server
func startServer(config *Config, client HTTPClient) {
	http.HandleFunc("/", createHandler(config, client))

	log.Printf("Starting server on :8080 (SYMBOL=%s, NDAYS=%d)", config.Symbol, config.NDays)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// createHandler creates the HTTP handler for the stock ticker endpoint
func createHandler(config *Config, client HTTPClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		data, avgClose, err := fetchStockData(config.Symbol, config.NDays, config.APIKey, client)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching stock data: %v", err), http.StatusInternalServerError)
			return
		}

		response := StockResponse{
			Symbol:       config.Symbol,
			Days:         config.NDays,
			AverageClose: avgClose,
			Data:         data,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// fetchStockData gets stock data from the Alpha Vantage API
func fetchStockData(symbol string, nDays int, apiKey string, client HTTPClient) ([]TimeSeriesData, float64, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?apikey=%s&function=TIME_SERIES_DAILY&symbol=%s", apiKey, symbol)
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var avResp AlphaVantageResponse
	if err := json.NewDecoder(resp.Body).Decode(&avResp); err != nil {
		return nil, 0, fmt.Errorf("error decoding response: %v", err)
	}

	if avResp.TimeSeries == nil {
		return nil, 0, fmt.Errorf("no time series data returned")
	}

	data, avgClose := processTimeSeries(avResp.TimeSeries, nDays)
	return data, avgClose, nil
}

// processTimeSeries processes the time series data from Alpha Vantage
func processTimeSeries(timeSeries map[string]map[string]interface{}, nDays int) ([]TimeSeriesData, float64) {
	var data []TimeSeriesData
	var totalClose float64

	var dates []string
	for date := range timeSeries {
		dates = append(dates, date)
	}
	
	limit := nDays
	if len(dates) < nDays {
		limit = len(dates)
	}

	for i := 0; i < limit && i < len(dates); i++ {
		date := dates[i]
		dayData := timeSeries[date]
		
		closePriceStr, ok := dayData["4. close"].(string)
		if !ok {
			continue
		}
		closePrice, err := strconv.ParseFloat(closePriceStr, 64)
		if err != nil {
			continue
		}

		openPriceStr, _ := dayData["1. open"].(string)
		openPrice, _ := strconv.ParseFloat(openPriceStr, 64)

		highPriceStr, _ := dayData["2. high"].(string)
		highPrice, _ := strconv.ParseFloat(highPriceStr, 64)

		lowPriceStr, _ := dayData["3. low"].(string)
		lowPrice, _ := strconv.ParseFloat(lowPriceStr, 64)

		volumeStr, _ := dayData["5. volume"].(string)
		volume, _ := strconv.ParseInt(volumeStr, 10, 64)

		data = append(data, TimeSeriesData{
			Date:       date,
			OpenPrice:  openPrice,
			HighPrice:  highPrice,
			LowPrice:   lowPrice,
			ClosePrice: closePrice,
			Volume:     volume,
		})

		totalClose += closePrice
	}

	var avgClose float64
	if len(data) > 0 {
		avgClose = totalClose / float64(len(data))
	}

	return data, avgClose
}