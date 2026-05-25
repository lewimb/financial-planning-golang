package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Timeouts per ML service recommendation (see ML_SERVICE_INTEGRATION.md).
const (
	analysisTimeout       = 5 * time.Second
	anomalyTimeout        = 10 * time.Second
	forecastTimeout       = 60 * time.Second
	insightsTimeout       = 10 * time.Second
	forecastStartTimeout  = 5 * time.Second
	forecastStatusTimeout = 5 * time.Second
)

// Client is an HTTP client for the ML service.
// Base URL is read from ML_SERVICE_URL env var; falls back to localhost:8000.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient constructs a Client. Call once at startup and reuse.
func NewClient() *Client {
	base := os.Getenv("ML_SERVICE_URL")
	if base == "" {
		base = "http://localhost:8000"
		log.Printf("WARNING: ML_SERVICE_URL not set — ML calls will target %s; set this env var in production", base)
	}
	return &Client{
		baseURL:    base,
		httpClient: &http.Client{},
	}
}

// post is the shared POST helper. It marshals body, sends the request within ctx,
// and decodes the JSON response into dst.
func (c *Client) post(ctx context.Context, path string, body interface{}, dst interface{}) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("ml: marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("ml: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("ML_API_KEY"); key != "" {
		req.Header.Set("X-Api-Key", key)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ml: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ml: unexpected status %d from %s", resp.StatusCode, path)
	}

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("ml: decode response: %w", err)
	}
	return nil
}

// Analysis calls POST /analysis with a 5 s timeout.
func (c *Client) Analysis(transactions []Transaction) (*AnalysisResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()

	var result AnalysisResponse
	if err := c.post(ctx, "/analysis", transactions, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Anomaly calls POST /anomaly with a 10 s timeout.
func (c *Client) Anomaly(transactions []Transaction) (*AnomalyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), anomalyTimeout)
	defer cancel()

	var result AnomalyResponse
	if err := c.post(ctx, "/anomaly", transactions, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Insights calls POST /insights with a 10 s timeout.
func (c *Client) Insights(transactions []Transaction) (*InsightsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), insightsTimeout)
	defer cancel()

	var result InsightsResponse
	if err := c.post(ctx, "/insights", transactions, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ForecastStart calls POST /forecast/start and returns a job ID immediately (202 Accepted).
// The ML service runs the forecast asynchronously; poll ForecastStatus for the result.
func (c *Client) ForecastStart(transactions []Transaction, periods int) (*ForecastJobResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), forecastStartTimeout)
	defer cancel()

	if periods <= 0 {
		periods = 30
	}
	if periods > 365 {
		periods = 365
	}

	payload, err := json.Marshal(transactions)
	if err != nil {
		return nil, fmt.Errorf("ml: marshal body: %w", err)
	}

	url := fmt.Sprintf("%s/forecast/start?periods=%d", c.baseURL, periods)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("ml: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if key := os.Getenv("ML_API_KEY"); key != "" {
		req.Header.Set("X-Api-Key", key)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ml: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("ml: unexpected status %d from /forecast/start", resp.StatusCode)
	}

	var result ForecastJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ml: decode response: %w", err)
	}
	return &result, nil
}

// ForecastStatus calls GET /forecast/status/{jobID} and returns the current job state.
func (c *Client) ForecastStatus(jobID string) (*ForecastStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), forecastStatusTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/forecast/status/%s", c.baseURL, jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("ml: build request: %w", err)
	}
	if key := os.Getenv("ML_API_KEY"); key != "" {
		req.Header.Set("X-Api-Key", key)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ml: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("ml: job %s not found", jobID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml: unexpected status %d from /forecast/status", resp.StatusCode)
	}

	var result ForecastStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ml: decode response: %w", err)
	}
	return &result, nil
}

// Forecast calls POST /forecast?periods=N with a 60 s timeout.
// periods is clamped to [1, 365]; defaults to 30 when zero.
func (c *Client) Forecast(transactions []Transaction, periods int) (*ForecastResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), forecastTimeout)
	defer cancel()

	if periods <= 0 {
		periods = 30
	}
	if periods > 365 {
		periods = 365
	}

	payload, err := json.Marshal(transactions)
	if err != nil {
		return nil, fmt.Errorf("ml: marshal body: %w", err)
	}

	url := fmt.Sprintf("%s/forecast?periods=%d", c.baseURL, periods)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("ml: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ml: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ml: unexpected status %d from /forecast", resp.StatusCode)
	}

	var result ForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("ml: decode response: %w", err)
	}
	return &result, nil
}
