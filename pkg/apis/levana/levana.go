package levana

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// FundingRate represents a single funding rate entry
type FundingRate struct {
	Timestamp string `json:"timestamp"`
	LongRate  string `json:"long_rate"`
	ShortRate string `json:"short_rate"`
}

// Client defines the funding rate API client
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient initializes a new funding rate API client with a custom HTTP client
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second} // Default timeout
	}
	return &Client{
		BaseURL: baseURL,
		HTTP:    httpClient,
	}
}

// FetchFundingRates retrieves funding rates from the Levana API with exponential backoff
func (c *Client) FetchFundingRates(market, startDate, endDate string) ([]FundingRate, error) {
	url := fmt.Sprintf("%s/funding-rates?market=%s&start_date=%s&end_date=%s", c.BaseURL, market, startDate, endDate)

	var rates []FundingRate

	// Define the operation to be retried
	operation := func() error {
		resp, err := c.HTTP.Get(url)
		if err != nil {
			return fmt.Errorf("error making GET request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("received non-OK status code: %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&rates); err != nil {
			return fmt.Errorf("error unmarshaling JSON: %w", err)
		}

		return nil
	}

	// Configure the exponential backoff parameters
	expBackoff := backoff.NewExponentialBackOff()

	// Retry the operation using the backoff strategy
	if err := backoff.Retry(operation, expBackoff); err != nil {
		return nil, fmt.Errorf("failed to fetch funding rates after retries: %w", err)
	}

	return rates, nil
}
