package perps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/margined-protocol/locust-core/pkg/contracts/base"
	"google.golang.org/grpc"
)

// QueryClient is the client API for Query service.
type QueryClient interface {
	MarketState(ctx context.Context, req *MarketStateRequest, opts ...grpc.CallOption) (*MarketStateResponse, error)
	Market(ctx context.Context, req *MarketRequest, opts ...grpc.CallOption) (*MarketResponse, error)
	Markets(ctx context.Context, req *MarketsRequest, opts ...grpc.CallOption) (*[]MarketResponse, error)
	Position(ctx context.Context, req *PositionRequest, opts ...grpc.CallOption) (*PositionResponse, error)
	Positions(ctx context.Context, req *PositionsRequest, opts ...grpc.CallOption) (*[]PositionResponse, error)
	PositionsByAccount(ctx context.Context, req *PositionsByAccountRequest, opts ...grpc.CallOption) (*PositionsByAccountResponse, error)
	RealizedPnlByAccountAndMarket(ctx context.Context, req *RealizedPnlRequest, opts ...grpc.CallOption) (*PnlAmounts, error)
	OpeningFee(ctx context.Context, req *OpeningFeeRequest, opts ...grpc.CallOption) (*TradingFee, error)
	PositionFees(ctx context.Context, req *PositionFeesRequest, opts ...grpc.CallOption) (*PositionFeesResponse, error)
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

// Generic query handler
func (q *queryClient) query(ctx context.Context, queryType string, req interface{}, resp interface{}, opts ...grpc.CallOption) error {
	rawQueryData, err := json.Marshal(map[string]interface{}{
		queryType: req,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal %s request: %w", queryType, err)
	}

	rawResponseData, err := q.baseQueryClient.QuerySmartContractState(ctx, q.address, rawQueryData, opts...)
	if err != nil {
		return fmt.Errorf("failed to query %s: %w", queryType, err)
	}

	if err := json.Unmarshal(rawResponseData, resp); err != nil {
		return fmt.Errorf("failed to unmarshal %s response: %w", queryType, err)
	}

	return nil
}

// Individual query implementations
func (q *queryClient) MarketState(ctx context.Context, req *MarketStateRequest, opts ...grpc.CallOption) (*MarketStateResponse, error) {
	var response MarketStateResponse
	err := q.query(ctx, "market_state", req, &response, opts...)
	return &response, err
}

func (q *queryClient) Market(ctx context.Context, req *MarketRequest, opts ...grpc.CallOption) (*MarketResponse, error) {
	var response MarketResponse
	err := q.query(ctx, "market", req, &response, opts...)
	return &response, err
}

func (q *queryClient) Markets(ctx context.Context, req *MarketsRequest, opts ...grpc.CallOption) (*[]MarketResponse, error) {
	var response []MarketResponse
	err := q.query(ctx, "markets", req, &response, opts...)
	return &response, err
}

func (q *queryClient) Position(ctx context.Context, req *PositionRequest, opts ...grpc.CallOption) (*PositionResponse, error) {
	var response PositionResponse
	err := q.query(ctx, "position", req, &response, opts...)
	return &response, err
}

func (q *queryClient) Positions(ctx context.Context, req *PositionsRequest, opts ...grpc.CallOption) (*[]PositionResponse, error) {
	var response []PositionResponse
	err := q.query(ctx, "positions", req, &response, opts...)
	return &response, err
}

func (q *queryClient) PositionsByAccount(ctx context.Context, req *PositionsByAccountRequest, opts ...grpc.CallOption) (*PositionsByAccountResponse, error) {
	var response PositionsByAccountResponse
	err := q.query(ctx, "positions_by_account", req, &response, opts...)
	return &response, err
}

func (q *queryClient) RealizedPnlByAccountAndMarket(ctx context.Context, req *RealizedPnlRequest, opts ...grpc.CallOption) (*PnlAmounts, error) {
	var response PnlAmounts
	err := q.query(ctx, "realized_pnl_by_account_and_market", req, &response, opts...)
	return &response, err
}

func (q *queryClient) OpeningFee(ctx context.Context, req *OpeningFeeRequest, opts ...grpc.CallOption) (*TradingFee, error) {
	var response TradingFee
	err := q.query(ctx, "opening_fee", req, &response, opts...)
	return &response, err
}

func (q *queryClient) PositionFees(ctx context.Context, req *PositionFeesRequest, opts ...grpc.CallOption) (*PositionFeesResponse, error) {
	var response PositionFeesResponse
	err := q.query(ctx, "position_fees", req, &response, opts...)
	return &response, err
}
