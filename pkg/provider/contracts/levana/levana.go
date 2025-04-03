package astroport

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cenkalti/backoff/v4"
	locustbackoff "github.com/margined-protocol/locust-core/pkg/backoff"
	providertypes "github.com/margined-protocol/locust-core/pkg/provider"
	"go.uber.org/zap"
)

// APIResponse represents the structure of the JSON response from the Astroport API.
type APIResponse map[string]map[string]float64

type SpotPriceResponse struct {
	PriceNotional  string  `json:"price_notional"`
	PriceUSD       string  `json:"price_usd"`
	PriceBase      string  `json:"price_base"`
	Timestamp      string  `json:"timestamp"`
	IsNotionalUSD  bool    `json:"is_notional_usd"`
	MarketType     string  `json:"market_type"`
	PublishTime    *string `json:"publish_time"`
	PublishTimeUSD *string `json:"publish_time_usd"`
}

func QuerySpotPrice(ctx context.Context, qc wasmdtypes.QueryClient, contractAddress string) (*SpotPriceResponse, error) {
	query := `{  "spot_price": {}}`

	queryBytes := []byte(query)
	req := wasmdtypes.QuerySmartContractStateRequest{Address: contractAddress, QueryData: queryBytes}

	resp, err := qc.SmartContractState(ctx, &req)
	if err != nil {
		return nil, err
	}

	var spotPriceResponse SpotPriceResponse
	if err := json.Unmarshal(resp.Data, &spotPriceResponse); err != nil {
		return nil, err
	}

	return &spotPriceResponse, nil
}

// Astroport implements a Provider that fetches prices using the Astroport API.
type Levana struct {
	wasmQueryClient *wasmdtypes.QueryClient
	logger          *zap.Logger
	contractAddress string
}

// Ensure Astroport implements the Provider interface
var _ providertypes.Provider = (*Levana)(nil)

// NewClient creates a new Astroport API provider.
func NewClient(client *wasmdtypes.QueryClient, contractAddress string, logger *zap.Logger) *Levana {
	return &Levana{
		wasmQueryClient: client,
		contractAddress: contractAddress,
		logger:          logger,
	}
}

// Name returns the name of the provider.
func (*Levana) Name() string {
	return "Astroport"
}

func (l *Levana) FetchPrice(ctx context.Context) (float64, error) {
	var price float64

	l.logger.Debug("contract address",
		zap.String("contractAddress", l.contractAddress),
	)

	// Prepare backoff strategy for retrying the request
	exponentialBackoff := locustbackoff.NewBackoff(ctx)

	retryableRequest := func() error {
		// Use QuerySimulation to get the price from the smart contract
		spotPriceResponse, err := QuerySpotPrice(ctx, *l.wasmQueryClient, l.contractAddress)
		if err != nil {
			l.logger.Error("Failed to query smart contract", zap.Error(err))
			return err
		}

		// Convert return_amount from string to float64
		price, err = strconv.ParseFloat(spotPriceResponse.PriceNotional, 64)
		if err != nil {
			l.logger.Error("Failed to convert return amount to float64", zap.Error(err))
			return err
		}

		l.logger.Info("Fetched price successfully", zap.Float64("price", price))

		// Return nil error to indicate success for retry logic
		return nil
	}

	// Retry the request with exponential backoff
	err := backoff.Retry(retryableRequest, exponentialBackoff)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch price after retries: %w", err)
	}

	// Successfully fetched and converted price
	return price, nil
}
