package milkyway

import (
	"context"
	"encoding/json"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the API for querying an astroport contract.
type QueryClient interface {
	QueryBatch(ctx context.Context, contractAddress string, batchID uint64, opts ...grpc.CallOption) (*BatchResponse, error)
	QueryUnstakeRequest(ctx context.Context, contractAddress, user string, opts ...grpc.CallOption) (*UnstakeRequestResponse, error)
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

func (q *queryClient) QueryBatch(ctx context.Context, contractAddress string, batchID uint64, opts ...grpc.CallOption) (*BatchResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"batch": map[string]any{
			"id": batchID,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var batchResponse BatchResponse
	if err := json.Unmarshal(rawResponseData, &batchResponse); err != nil {
		return nil, err
	}

	return &batchResponse, nil
}

func (q *queryClient) QueryUnstakeRequest(ctx context.Context, contractAddress, user string, opts ...grpc.CallOption) (*UnstakeRequestResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{"unstake_requests": map[string]any{
		"user": user,
	}})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var unstakeResponse UnstakeRequestResponse
	if err := json.Unmarshal(rawResponseData, &unstakeResponse.Requests); err != nil {
		return nil, err
	}

	return &unstakeResponse, nil
}
