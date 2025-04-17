package lpp

import (
	"context"
	"encoding/json"
	"math/big"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// CoinDTO represents a coin with denom and amount
type CoinDTO struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// Query client interface definitions
type QueryClient interface {
	// LppBalance returns the total balance of the LPP
	LppBalance(ctx context.Context, req *LppBalanceRequest) (*LppBalanceResponse, error)
	// Price returns the current nLPN/LPN exchange rate
	Price(ctx context.Context, req *PriceRequest) (*PriceResponse, error)
	// Balance returns a user's nLPN balance
	Balance(ctx context.Context, req *BalanceRequest) (*BalanceResponse, error)
	// DepositCapacity returns the maximum deposit possible due to utilization constraints
	DepositCapacity(ctx context.Context, req *DepositCapacityRequest) (*DepositCapacityResponse, error)
	// Quote returns information about potential loan interest rates
	Quote(ctx context.Context, req *QuoteRequest) (*QuoteResponse, error)
}

// QueryClientImpl implements the QueryClient interface
type QueryClientImpl struct {
	grpcConn  *grpc.ClientConn
	contract  string
	wasmQuery wasmtypes.QueryClient
}

// NewQueryClient creates a new query client for the LPP contract
func NewQueryClient(conn *grpc.ClientConn, contract string) QueryClient {
	return &QueryClientImpl{
		grpcConn:  conn,
		contract:  contract,
		wasmQuery: wasmtypes.NewQueryClient(conn),
	}
}

// LppBalanceRequest is the request for LppBalance query
type LppBalanceRequest struct{}

// LppBalanceResponse is the response from LppBalance query
type LppBalanceResponse struct {
	Balance           CoinDTO `json:"balance"`
	TotalPrincipalDue CoinDTO `json:"total_principal_due"`
	TotalInterestDue  CoinDTO `json:"total_interest_due"`
	BalanceNlpn       CoinDTO `json:"balance_nlpn"`
}

// LppBalance queries the total balance in the LPP
func (q *QueryClientImpl) LppBalance(ctx context.Context, req *LppBalanceRequest) (*LppBalanceResponse, error) {
	queryMsg := struct {
		LppBalance struct{} `json:"lpp_balance"`
	}{
		LppBalance: struct{}{},
	}

	res, err := q.queryContract(ctx, queryMsg)
	if err != nil {
		return nil, err
	}

	var response LppBalanceResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, status.Error(codes.Internal, "failed to unmarshal response")
	}

	return &response, nil
}

// PriceRequest is the request for Price query
type PriceRequest struct{}

// PriceResponse is the response from Price query
type PriceResponse struct {
	Amount      string `json:"amount"`
	AmountQuote string `json:"amount_quote"`
}

// Price queries the current price of nLPN in terms of LPN
func (q *QueryClientImpl) Price(ctx context.Context, req *PriceRequest) (*PriceResponse, error) {
	queryMsg := struct {
		Price struct{} `json:"price"`
	}{
		Price: struct{}{},
	}

	res, err := q.queryContract(ctx, queryMsg)
	if err != nil {
		return nil, err
	}

	// Parse the nested API response
	var apiResponse struct {
		Data struct {
			Amount struct {
				Amount string `json:"amount"`
			} `json:"amount"`
			AmountQuote struct {
				Amount string `json:"amount"`
			} `json:"amount_quote"`
		} `json:"data"`
	}

	if err := json.Unmarshal(res, &apiResponse); err != nil {
		return nil, status.Error(codes.Internal, "failed to unmarshal response")
	}

	// Map to the expected flat response structure
	response := &PriceResponse{
		Amount:      apiResponse.Data.Amount.Amount,
		AmountQuote: apiResponse.Data.AmountQuote.Amount,
	}

	return response, nil
}

// BalanceRequest is the request for Balance query
type BalanceRequest struct {
	Address string `json:"address"`
}

// BalanceResponse is the response from Balance query
type BalanceResponse struct {
	Balance big.Int `json:"balance"`
}

// Balance queries a user's nLPN balance
func (q *QueryClientImpl) Balance(ctx context.Context, req *BalanceRequest) (*BalanceResponse, error) {
	queryMsg := struct {
		Balance struct {
			Address string `json:"address"`
		} `json:"balance"`
	}{
		Balance: struct {
			Address string `json:"address"`
		}{
			Address: req.Address,
		},
	}

	res, err := q.queryContract(ctx, queryMsg)
	if err != nil {
		return nil, err
	}

	var response BalanceResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, status.Error(codes.Internal, "failed to unmarshal response")
	}

	return &response, nil
}

// DepositCapacityRequest is the request for DepositCapacity query
type DepositCapacityRequest struct{}

// DepositCapacityResponse is the response from DepositCapacity query
type DepositCapacityResponse struct {
	Capacity *CoinDTO `json:"capacity"`
}

// DepositCapacity queries the maximum deposit possible
func (q *QueryClientImpl) DepositCapacity(ctx context.Context, req *DepositCapacityRequest) (*DepositCapacityResponse, error) {
	queryMsg := struct {
		DepositCapacity struct{} `json:"deposit_capacity"`
	}{
		DepositCapacity: struct{}{},
	}

	res, err := q.queryContract(ctx, queryMsg)
	if err != nil {
		return nil, err
	}

	var response DepositCapacityResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, status.Error(codes.Internal, "failed to unmarshal response")
	}

	return &response, nil
}

// QuoteRequest is the request for Quote query
type QuoteRequest struct {
	Amount CoinDTO `json:"amount"`
}

// QuoteResponse is the response from Quote query
type QuoteResponse struct {
	QuoteInterestRate *struct {
		Value string `json:"value"`
	} `json:"quote_interest_rate,omitempty"`
	NoLiquidity *struct{} `json:"no_liquidity,omitempty"`
}

// Quote queries the potential interest rate for a loan
func (q *QueryClientImpl) Quote(ctx context.Context, req *QuoteRequest) (*QuoteResponse, error) {
	queryMsg := struct {
		Quote struct {
			Amount CoinDTO `json:"amount"`
		} `json:"quote"`
	}{
		Quote: struct {
			Amount CoinDTO `json:"amount"`
		}{
			Amount: req.Amount,
		},
	}

	res, err := q.queryContract(ctx, queryMsg)
	if err != nil {
		return nil, err
	}

	var response QuoteResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, status.Error(codes.Internal, "failed to unmarshal response")
	}

	return &response, nil
}

// queryContract is a helper function for querying a CosmWasm contract
func (q *QueryClientImpl) queryContract(ctx context.Context, queryMsg interface{}) ([]byte, error) {
	queryData, err := json.Marshal(queryMsg)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to marshal query")
	}

	res, err := q.wasmQuery.SmartContractState(
		ctx,
		&wasmtypes.QuerySmartContractStateRequest{
			Address:   q.contract,
			QueryData: queryData,
		},
	)
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
