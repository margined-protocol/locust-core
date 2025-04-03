package drop

import (
	"context"
	"encoding/json"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// DropQuerier defines the interface for querying the Drop contract.
type QueryClient interface {
	QueryTokens(ctx context.Context, contractAddress, owner string, opts ...grpc.CallOption) (*TokensResponse, error)
	QueryNftInfo(ctx context.Context, contractAddress, tokenID string, opts ...grpc.CallOption) (*NftInfoResponse, error)
	QueryUnbondBatch(ctx context.Context, contractAddress, bondID string, opts ...grpc.CallOption) (*UnbondBatchResponse, error)
	Close() error
}

type queryClient struct {
	baseQueryClient base.QueryClient
	cc              *grpc.ClientConn
}

var _ QueryClient = (*queryClient)(nil)

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

// QueryTokens returns NFT tokens for an account
func (q *queryClient) QueryTokens(ctx context.Context, contractAddress, owner string, opts ...grpc.CallOption) (*TokensResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"tokens": map[string]any{
			"owner": owner,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var tokensResponse TokensResponse
	if err := json.Unmarshal(rawResponseData, &tokensResponse); err != nil {
		return nil, err
	}

	return &tokensResponse, nil
}

// QueryNftInfo returns the NFT info for a given token ID
func (q *queryClient) QueryNftInfo(ctx context.Context, contractAddress, tokenID string, opts ...grpc.CallOption) (*NftInfoResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"nft_info": map[string]any{
			"token_id": tokenID,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var nftInfoResponse NftInfoResponse
	if err := json.Unmarshal(rawResponseData, &nftInfoResponse); err != nil {
		return nil, err
	}

	return &nftInfoResponse, nil
}

// QueryUnbondBatch returns the unbond batch info for a given batch ID
func (q *queryClient) QueryUnbondBatch(ctx context.Context, contractAddress, batchID string, opts ...grpc.CallOption) (*UnbondBatchResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"unbond_batch": map[string]any{
			"batch_id": batchID,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var unbondBatchResponse UnbondBatchResponse
	if err := json.Unmarshal(rawResponseData, &unbondBatchResponse); err != nil {
		return nil, err
	}

	return &unbondBatchResponse, nil
}
