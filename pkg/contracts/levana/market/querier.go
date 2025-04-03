package market

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	Close() error
	GetDeferredExecID(ctx context.Context, contractAddress, id string, opts ...grpc.CallOption) (*GetDeferredExecIDResponse, error)
	QueryLastCrankCompleted(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (int64, error)
	NftProxy(ctx context.Context, contractAddress, owner string, startAfter *string, limit *int, opts ...grpc.CallOption) (*NftProxyResponse, error)
	PositionActionHistory(ctx context.Context, contractAddress, positionID string, opts ...grpc.CallOption) (*PositionActionHistoryResponse, error)
	Positions(ctx context.Context, contractAddress string, positionIDs []string, opts ...grpc.CallOption) (*PositionsResponse, error)
	Status(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*StatusResponse, error)
}

// PositionsResponse defines the response structure for open positions
type NftProxyResponse struct {
	Tokens []string `json:"tokens"`
}

type queryClient struct {
	baseQueryClient base.QueryClient
	cc              *grpc.ClientConn
}

var _ QueryClient = (*queryClient)(nil)

// NewQueryClient creates a new QueryClient with an optional backoff strategy
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

// Positions queries active and closed positions by position IDs
func (q *queryClient) Positions(ctx context.Context, contractAddress string, positionIDs []string, opts ...grpc.CallOption) (*PositionsResponse, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"positions": map[string]interface{}{
			"position_ids": positionIDs,
		},
	})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var response PositionsResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// Status queries the contract for the status
func (q *queryClient) Status(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (*StatusResponse, error) {
	rawQueryData, err := json.Marshal(map[string]any{"status": struct{}{}})
	if err != nil {
		return nil, err
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, err
	}

	var response StatusResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// QueryLastCrankCompleted queries the raw contract state to get the last crank completion time
func (q *queryClient) QueryLastCrankCompleted(ctx context.Context, contractAddress string, opts ...grpc.CallOption) (int64, error) {
	key := []byte("f")

	rawResponseData, err := q.baseQueryClient.QueryRawContractState(ctx, contractAddress, key, opts...)
	if err != nil {
		return 0, err
	}

	rawString := strings.Trim(string(rawResponseData), "\"")

	timestamp, err := strconv.ParseInt(rawString, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp: %v", err)
	}

	return timestamp, nil
}

// NftProxy queries NFTs for a particular user, supporting pagination (start_after, limit)
func (q *queryClient) NftProxy(ctx context.Context, contractAddress, owner string, startAfter *string, limit *int, opts ...grpc.CallOption) (*NftProxyResponse, error) {
	// Construct base query
	query := map[string]any{
		"nft_proxy": map[string]any{
			"nft_msg": map[string]any{
				"tokens": map[string]any{
					"owner": owner,
				},
			},
		},
	}

	// Add optional parameters if provided
	if startAfter != nil {
		query["nft_proxy"].(map[string]any)["nft_msg"].(map[string]any)["tokens"].(map[string]any)["start_after"] = *startAfter
	}
	if limit != nil {
		query["nft_proxy"].(map[string]any)["nft_msg"].(map[string]any)["tokens"].(map[string]any)["limit"] = *limit
	}

	// Marshal to JSON
	rawQueryData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("error marshalling query: %w", err)
	}

	// Query contract
	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("error querying contract: %w", err)
	}

	// Parse response
	var nftProxyResponse NftProxyResponse
	if err := json.Unmarshal(rawResponseData, &nftProxyResponse); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &nftProxyResponse, nil
}

// PositionActionHistory queries the contract for position action history
func (q *queryClient) PositionActionHistory(ctx context.Context, contractAddress, positionID string, opts ...grpc.CallOption) (*PositionActionHistoryResponse, error) {
	// Construct query payload
	query := map[string]any{
		"position_action_history": map[string]any{
			"id": positionID,
		},
	}

	// Marshal to JSON
	rawQueryData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("error marshalling query: %w", err)
	}

	// Query contract
	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("error querying contract: %w", err)
	}

	// Parse response
	var response PositionActionHistoryResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	return &response, nil
}

func (q *queryClient) GetDeferredExecID(ctx context.Context, contractAddress, id string, opts ...grpc.CallOption) (*GetDeferredExecIDResponse, error) {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		"get_deferred_exec": map[string]string{
			"id": id,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error marshalling query: %w", err)
	}

	// Query the contract
	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, contractAddress, rawQueryData, opts...)
	if err != nil {
		return nil, fmt.Errorf("error querying contract: %w", err)
	}

	// Debug log raw response
	fmt.Println("Raw Deferred Exec Response:", string(rawResponseData))

	// Attempt to unmarshal
	var response GetDeferredExecIDResponse
	if err := json.Unmarshal(rawResponseData, &response); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %w", err)
	}

	// Handle "not_found" as a valid state
	if response.NotFound != nil {
		fmt.Println("Deferred exec ID not found:", id)
		return &response, nil // ✅ No error, just `Found=nil`
	}

	// If "found" is missing, return an empty response (just in case)
	if response.Found == nil {
		return &response, nil
	}

	return &response, nil
}
