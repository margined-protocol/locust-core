package fund

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// QueryClient defines the interface for querying the Fund contract.
type QueryClient interface {
	QueryWithdrawableAssets(ctx context.Context, contractAddress string, opts ...grpc.CallOption) ([]sdk.Coin, error)
	QueryState(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*StateResponse, error)
	QueryPendingRedemptions(ctx context.Context, contractAddress string, limit *uint64, opts ...grpc.CallOption) ([]Redemption, error)
	QueryConvertToAssets(ctx context.Context, contractAddress string, amount sdkmath.Int, opts ...grpc.CallOption) (sdkmath.Int, error)
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

// QueryWithdrawableAssets queries the withdrawable assets for the vault
func (q *queryClient) QueryWithdrawableAssets(ctx context.Context, contractAddress string, opts ...grpc.CallOption) ([]sdk.Coin, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"vault_extension": map[string]any{
			"vaultenator": map[string]any{
				"withdrawable_amount": map[string]any{},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var coins []sdk.Coin
	if err := json.Unmarshal(rawResponseData, &coins); err != nil {
		return nil, err
	}

	return coins, nil
}

// QueryState queries the current state of the vault
func (q *queryClient) QueryState(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*StateResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{
		"vault_extension": map[string]any{
			"vaultenator": map[string]any{
				"state": map[string]any{},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal state query: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query state: %w", err)
	}

	var state StateResponse
	if err := json.Unmarshal(rawResponseData, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state response: %w", err)
	}

	return &state, nil
}

// QueryPendingRedemptions queries pending redemptions from the vault
func (q *queryClient) QueryPendingRedemptions(ctx context.Context, contractAddress string, limit *uint64, opts ...grpc.CallOption) ([]Redemption, error) {
	query := map[string]any{
		"pending_redemptions": map[string]any{},
	}

	// Add limit if provided
	if limit != nil {
		query["pending_redemptions"] = map[string]any{
			"limit": *limit,
		}
	}

	rawQueryData, err := json.Marshal(map[string]any{
		"vault_extension": map[string]any{
			"vaultenator": query,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pending redemptions query: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending redemptions: %w", err)
	}

	var redemptions []Redemption
	if err := json.Unmarshal(rawResponseData, &redemptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal redemptions response: %w", err)
	}

	return redemptions, nil
}

// QueryConvertToAssets converts a given amount of shares to the equivalent asset amount
func (q *queryClient) QueryConvertToAssets(ctx context.Context, contractAddress string, amount sdkmath.Int, opts ...grpc.CallOption) (sdkmath.Int, error) {
	// Construct the query
	queryData := map[string]any{
		"convert_to_assets": map[string]any{
			"amount": amount.String(),
		},
	}

	rawQueryData, err := json.Marshal(queryData)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to marshal convert_to_assets query: %w", err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to query convert_to_assets: %w", err)
	}

	// Response is a string representation of the Uint128 value
	var assetAmountStr string
	if err := json.Unmarshal(rawResponseData, &assetAmountStr); err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to unmarshal convert_to_assets response: %w", err)
	}

	// Convert string to sdkmath.Int
	assetAmount, ok := sdkmath.NewIntFromString(assetAmountStr)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("failed to convert asset amount string to Int: %s", assetAmountStr)
	}

	return assetAmount, nil
}

// Response structs based on the Rust contract definitions

// StateResponse represents the vault state information
type StateResponse struct {
	IsOpen                bool       `json:"is_open"`
	IsPaused              bool       `json:"is_paused"`
	LastPause             string     `json:"last_pause"`
	LastClaim             string     `json:"last_claim"`
	TotalStakedTokens     string     `json:"total_staked_tokens"`
	TotalWithdrawnTokens  []sdk.Coin `json:"total_withdrawn_tokens"`
	PendingManagementFees []sdk.Coin `json:"pending_management_fees"`
}

// Redemption represents a pending redemption request
type Redemption struct {
	User      string `json:"user"`
	Amount    string `json:"amount"`
	Timestamp uint64 `json:"timestamp"`
}

// Helper methods for the response structs

// GetTotalStakedTokens returns the total staked tokens as an Int
func (s *StateResponse) GetTotalStakedTokens() (sdkmath.Int, error) {
	totalStakedTokens, ok := sdkmath.NewIntFromString(s.TotalStakedTokens)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("failed to convert total staked tokens to Int")
	}
	return totalStakedTokens, nil
}

// GetAmount returns the amount as an Int
func (r *Redemption) GetAmount() (sdkmath.Int, error) {
	amount, ok := sdkmath.NewIntFromString(r.Amount)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("failed to convert amount to Int")
	}
	return amount, nil
}

// GetTimestamp returns the redemption timestamp as a Time
func (r *Redemption) GetTimestamp() time.Time {
	return time.Unix(int64(r.Timestamp), 0)
}
