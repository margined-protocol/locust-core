package redbank

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	MarketV2(ctx context.Context, req *MarketV2Request, opts ...grpc.CallOption) (*MarketV2Response, error)
	Close() error
}

type queryClient struct {
	baseQueryClient base.QueryClient
	cc              *grpc.ClientConn
	address         string
}

var _ QueryClient = (*queryClient)(nil)

// NewQueryClient creates a new QueryClient
func NewQueryClient(conn *grpc.ClientConn, contractAddress string) QueryClient {
	baseQueryClient := base.NewQueryClient(conn)
	return &queryClient{
		baseQueryClient: *baseQueryClient,
		cc:              conn,
		address:         contractAddress,
	}
}

// Close closes the gRPC connection to the server
func (q *queryClient) Close() error {
	return q.cc.Close()
}

// MarketV2 queries the contract for the account kind
func (q *queryClient) MarketV2(ctx context.Context, req *MarketV2Request, opts ...grpc.CallOption) (*MarketV2Response, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"market_v2": req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal MarketV2 request: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query account kind: %w", err)
	}

	var response MarketV2Response
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MarketV2 response: %w", err)
	}

	return &response, nil
}

// Request and Response Structs
type MarketV2Request struct {
	Denom string `json:"denom"`
}

// MarketV2Response represents the extended market response structure.
type MarketV2Response struct {
	CollateralTotalAmount string            `json:"collateral_total_amount"`
	DebtTotalAmount       string            `json:"debt_total_amount"`
	UtilizationRate       string            `json:"utilization_rate"`
	Denom                 string            `json:"denom"`
	ReserveFactor         string            `json:"reserve_factor"`
	InterestRateModel     InterestRateModel `json:"interest_rate_model"`
	BorrowIndex           string            `json:"borrow_index"`
	LiquidityIndex        string            `json:"liquidity_index"`
	BorrowRate            string            `json:"borrow_rate"`
	LiquidityRate         string            `json:"liquidity_rate"`
	IndexesLastUpdated    int64             `json:"indexes_last_updated"`
	CollateralTotalScaled string            `json:"collateral_total_scaled"`
	DebtTotalScaled       string            `json:"debt_total_scaled"`
}
