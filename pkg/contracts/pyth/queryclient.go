package pyth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cenkalti/backoff/v4"
	locustbackoff "github.com/margined-protocol/locust-core/pkg/backoff"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base API URL (constant)
const baseURL = "https://hermes.pyth.network/v2/updates/price/latest?encoding=base64"

type Binary struct {
	Encoding string   `json:"encoding"` // Encoding type (e.g., "hex", "base64")
	Data     []string `json:"data"`     // Raw binary data (as hex/base64)
}

type Price struct {
	Price       string `json:"price"`
	Confidence  string `json:"conf"`
	Exponent    int    `json:"expo"`
	PublishTime int64  `json:"publish_time"`
}

type Metadata struct {
	Slot               int64 `json:"slot"`
	ProofAvailableTime int64 `json:"proof_available_time"`
	PrevPublishTime    int64 `json:"prev_publish_time"`
}

type Parsed struct {
	ID       string   `json:"id"`
	Price    Price    `json:"price"`
	EmaPrice Price    `json:"ema_price"`
	Metadata Metadata `json:"metadata"`
}

type PriceResponse struct {
	Binary Binary   `json:"binary"`
	Parsed []Parsed `json:"parsed"`
}

// PythClient is the client API for Pyth service
type QueryClient interface {
	LatestPrice(ctx context.Context, id string) (*PriceResponse, error)
	QueryGetUpdatedFee(ctx context.Context, qc wasmdtypes.QueryClient, contractAddress string, hexData string) (sdk.Coins, error)
}

type queryClient struct {
	client *http.Client
}

var _ QueryClient = (*queryClient)(nil)

// NewQueryClient creates a new Pyth Query Client
func NewQueryClient(client *http.Client) QueryClient {
	return &queryClient{
		client: client,
	}
}

// Name returns the name of the Query Client.
func (*queryClient) Name() string {
	return "Pyth"
}

// LatestPrice fetches the latest price from the Pyth API
func (q *queryClient) LatestPrice(ctx context.Context, id string) (*PriceResponse, error) {
	var apiResponse PriceResponse

	// Construct the full URL by appending the ID to the base URL
	apiURL := fmt.Sprintf("%s&ids%%5B%%5D=%s", baseURL, id)
	fmt.Println(apiURL)

	exponentialBackoff := locustbackoff.NewBackoff(ctx)

	retryableRequest := func() error {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return err
		}

		resp, err := q.client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			return err
		}

		return nil
	}

	err := backoff.Retry(retryableRequest, exponentialBackoff)
	if err != nil {
		return nil, err
	}

	// Ensure there is at least one parsed entry
	if len(apiResponse.Parsed) == 0 {
		return nil, fmt.Errorf("no parsed data found in Pyth response")
	}

	return &apiResponse, nil
}

func (*queryClient) QueryGetUpdatedFee(ctx context.Context, qc wasmdtypes.QueryClient, contractAddress string, base64Data string) (sdk.Coins, error) {
	query := fmt.Sprintf(`{
        "get_update_fee": {
            "vaas": ["%s"]
        }
    }`, base64Data)

	queryBytes := []byte(query)
	req := wasmdtypes.QuerySmartContractStateRequest{Address: contractAddress, QueryData: queryBytes}

	var res *wasmdtypes.QuerySmartContractStateResponse
	var err error
	exponentialBackoff := locustbackoff.NewBackoff(ctx)

	// Retry logic using exponential backoff
	retryableRequest := func() error {
		res, err = qc.SmartContractState(ctx, &req)
		return err
	}

	if err := backoff.Retry(retryableRequest, exponentialBackoff); err != nil {
		return nil, err
	}

	// Unmarshal into a single sdk.Coin
	var coin sdk.Coin
	if err := json.Unmarshal(res.Data, &coin); err != nil {
		return nil, err
	}

	// Convert the single sdk.Coin into sdk.Coins
	return sdk.Coins{coin}, nil
}
