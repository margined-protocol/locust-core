package steak

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the API for querying an astroport contract.
type QueryClient interface {
	ExchangeRateBackbone(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*float64, error)
	ExchangeRateEris(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*float64, error)
	Close() error
}

type queryClient struct {
	baseQueryClient base.QueryClient
	cc              *grpc.ClientConn
}

var _ QueryClient = (*queryClient)(nil)

// NewQueryClient creates a new QueryClient
func NewQueryClient(conn *grpc.ClientConn) QueryClient {
	baseQueryClient := base.NewQueryClient(conn)
	return &queryClient{
		baseQueryClient: *baseQueryClient,
		cc:              conn,
	}
}

// Close closes the gRPC connection to the server
func (q *queryClient) Close() error {
	return q.cc.Close()
}

// State performs a binary search to find the maximum offer amount to achieve a target price.
func (q *queryClient) ExchangeRateEris(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*float64, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"state": map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var state ErisStateResponse
	if err := json.Unmarshal(rawResponseData, &state); err != nil {
		return nil, err
	}

	exchangeRate, err := strconv.ParseFloat(state.ExchangeRate, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting exchangeRate to float64: %w", err)
	}

	return &exchangeRate, nil
}

// State performs a binary search to find the maximum offer amount to achieve a target price.
func (q *queryClient) ExchangeRateBackbone(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*float64, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"state": map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var state BackboneStateResponse
	if err := json.Unmarshal(rawResponseData, &state); err != nil {
		return nil, err
	}

	exchangeRate, err := strconv.ParseFloat(state.ExchangeRate, 64)
	if err != nil {
		return nil, fmt.Errorf("error converting exchangeRate to float64: %w", err)
	}

	return &exchangeRate, nil
}
