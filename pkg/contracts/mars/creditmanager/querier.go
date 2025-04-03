package creditmanager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	AccountKind(ctx context.Context, req *AccountKindRequest, opts ...grpc.CallOption) (*string, error)
	Accounts(ctx context.Context, req *AccountsRequest, opts ...grpc.CallOption) (*[]Account, error)
	Config(ctx context.Context, opts ...grpc.CallOption) (*ConfigResponse, error)
	VaultUtilization(ctx context.Context, req *VaultUtilizationRequest, opts ...grpc.CallOption) (*VaultUtilizationResponse, error)
	Positions(ctx context.Context, req *PositionsRequest, opts ...grpc.CallOption) (*PositionsResponse, error)
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
	return &queryClient{
		baseQueryClient: base.QueryClient{},
		cc:              conn,
		address:         contractAddress,
	}
}

// Close closes the gRPC connection to the server
func (q *queryClient) Close() error {
	return q.cc.Close()
}

// AccountKind queries the contract for the account kind
func (q *queryClient) AccountKind(ctx context.Context, req *AccountKindRequest, opts ...grpc.CallOption) (*string, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"account_kind": req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal AccountKind request: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query account kind: %w", err)
	}

	response := string(rawResponseData)
	return &response, nil
}

// Accounts queries the contract for accounts
func (q *queryClient) Accounts(ctx context.Context, req *AccountsRequest, opts ...grpc.CallOption) (*[]Account, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"accounts": req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Accounts request: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %w", err)
	}

	var response []Account
	// var response AccountsResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Accounts response: %w", err)
	}

	return &response, nil
}

// Config queries the contract for configuration
func (q *queryClient) Config(ctx context.Context, opts ...grpc.CallOption) (*ConfigResponse, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"config": map[string]interface{}{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Config request: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query config: %w", err)
	}

	var response ConfigResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Config response: %w", err)
	}

	return &response, nil
}

// Positions fetches the `Positions` details for a given account ID.
func (q *queryClient) Positions(ctx context.Context, req *PositionsRequest, opts ...grpc.CallOption) (*PositionsResponse, error) {
	// Build the query payload
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"positions": req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Positions request: %w", err)
	}

	// Perform the query
	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}

	// Parse the response
	var response PositionsResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Positions response: %w", err)
	}

	return &response, nil
}

// VaultUtilization queries the contract for vault utilization
func (q *queryClient) VaultUtilization(ctx context.Context, req *VaultUtilizationRequest, opts ...grpc.CallOption) (*VaultUtilizationResponse, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"vault_utilization": req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal VaultUtilization request: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query vault utilization: %w", err)
	}

	var response VaultUtilizationResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal VaultUtilization response: %w", err)
	}

	return &response, nil
}

// Request and Response Structs
type AccountKindRequest struct {
	AccountID string `json:"account_id"`
}

type AccountsRequest struct {
	Owner      string  `json:"owner"`
	StartAfter *string `json:"start_after,omitempty"`
	Limit      *uint32 `json:"limit,omitempty"`
}

type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

type VaultUtilizationRequest struct {
	Vault string `json:"vault"`
}

type VaultUtilizationResponse struct {
	Vault       string `json:"vault"`
	Utilization Coin   `json:"utilization"`
}

type ConfigResponse struct {
	Ownership    string `json:"ownership"`
	AccountNFT   string `json:"account_nft"`
	RedBank      string `json:"red_bank"`
	Incentives   string `json:"incentives"`
	Oracle       string `json:"oracle"`
	Params       string `json:"params"`
	Perps        string `json:"perps"`
	Swapper      string `json:"swapper"`
	Zapper       string `json:"zapper"`
	Health       string `json:"health_contract"`
	KeeperConfig string `json:"keeper_fee_config"`
}

type Account struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
