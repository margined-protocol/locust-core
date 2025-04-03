package cw20

import (
	"context"
	"encoding/json"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient defines the interface for querying a CW20 contract.
type QueryClient interface {
	QueryBalance(ctx context.Context, contractAddress, owner string, opts ...grpc.CallOption) (*BalanceResponse, error)
	QueryTokenInfo(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*TokenInfoResponse, error)
	QueryAllowance(ctx context.Context, contractAddress, owner, spender string, opts ...grpc.CallOption) (*AllowanceResponse, error)
	Close() error
}

type queryClient struct {
	baseQueryClient base.QueryClient
	cc              *grpc.ClientConn
}

var _ QueryClient = (*queryClient)(nil)

// NewQueryClient creates a new CW20 QueryClient
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

// QueryBalance retrieves the CW20 balance of an address
func (q *queryClient) QueryBalance(ctx context.Context, contractAddress, owner string, opts ...grpc.CallOption) (*BalanceResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"balance": map[string]any{
			"address": owner,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var balanceResponse BalanceResponse
	if err := json.Unmarshal(rawResponseData, &balanceResponse); err != nil {
		return nil, err
	}

	return &balanceResponse, nil
}

// QueryTokenInfo retrieves CW20 token details like name, symbol, decimals, and total supply
func (q *queryClient) QueryTokenInfo(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*TokenInfoResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"token_info": struct{}{},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var tokenInfoResponse TokenInfoResponse
	if err := json.Unmarshal(rawResponseData, &tokenInfoResponse); err != nil {
		return nil, err
	}

	return &tokenInfoResponse, nil
}

// QueryAllowance retrieves the amount a spender is allowed to withdraw from an owner's account
func (q *queryClient) QueryAllowance(ctx context.Context, contractAddress, owner, spender string, opts ...grpc.CallOption) (*AllowanceResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"allowance": map[string]any{
			"owner":   owner,
			"spender": spender,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var allowanceResponse AllowanceResponse
	if err := json.Unmarshal(rawResponseData, &allowanceResponse); err != nil {
		return nil, err
	}

	return &allowanceResponse, nil
}
