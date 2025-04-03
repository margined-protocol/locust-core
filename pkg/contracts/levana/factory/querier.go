package factory

import (
	"context"
	"encoding/json"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	MarketInfo(ctx context.Context, req *MarketInfoRequest, opts ...grpc.CallOption) (*MarketInfoResponse, error)
	Markets(ctx context.Context, opts ...grpc.CallOption) (*MarketsResponse, error)
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

// MarketInfo queries the contract for market info
func (q *queryClient) MarketInfo(ctx context.Context, req *MarketInfoRequest, opts ...grpc.CallOption) (*MarketInfoResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{"market_info": req})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var response MarketInfoResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Markets queries the contract for the list of markets
func (q *queryClient) Markets(ctx context.Context, opts ...grpc.CallOption) (*MarketsResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"markets": map[string]any{
			"limit": 100,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var response MarketsResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
