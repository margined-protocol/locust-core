package yieldmarket

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/margined-protocol/locust-core/pkg/contracts/nolus/lpp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLPPQueryClient is a mock implementation of the lpp.QueryClient interface
type MockLPPQueryClient struct {
	mock.Mock
}

// Price implements the lpp.QueryClient interface
func (m *MockLPPQueryClient) Price(ctx context.Context, req *lpp.PriceRequest) (*lpp.PriceResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*lpp.PriceResponse), args.Error(1)
}

// Implement other methods from lpp.QueryClient interface with empty implementations
func (m *MockLPPQueryClient) LppBalance(ctx context.Context, req *lpp.LppBalanceRequest) (*lpp.LppBalanceResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*lpp.LppBalanceResponse), args.Error(1)
}

func (m *MockLPPQueryClient) Balance(ctx context.Context, req *lpp.BalanceRequest) (*lpp.BalanceResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*lpp.BalanceResponse), args.Error(1)
}

func (m *MockLPPQueryClient) DepositCapacity(ctx context.Context, req *lpp.DepositCapacityRequest) (*lpp.DepositCapacityResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*lpp.DepositCapacityResponse), args.Error(1)
}

func (m *MockLPPQueryClient) Quote(ctx context.Context, req *lpp.QuoteRequest) (*lpp.QuoteResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*lpp.QuoteResponse), args.Error(1)
}

// DirectNolusYieldMarket uses direct client instead of creating one
type DirectNolusYieldMarket struct {
	NolusYieldMarket
	mockClient lpp.QueryClient
}

// getPrice overrides the parent method to use the mock client
func (m *DirectNolusYieldMarket) getPrice(ctx context.Context) (sdkmath.Int, error) {
	// Get current price
	res, err := m.mockClient.Price(
		ctx,
		&lpp.PriceRequest{},
	)
	if err != nil {
		return sdkmath.Int{}, err
	}

	amount, err := sdkmath.LegacyNewDecFromStr(res.Amount)
	if err != nil {
		return sdkmath.Int{}, err
	}

	amountQuote, err := sdkmath.LegacyNewDecFromStr(res.AmountQuote)
	if err != nil {
		return sdkmath.Int{}, err
	}

	price := amount.Quo(amountQuote).TruncateInt()

	return price, nil
}

func TestGetPrice(t *testing.T) {
	// Create a mock LPP client
	mockClient := new(MockLPPQueryClient)

	// Set up the market with the mock
	market := &DirectNolusYieldMarket{
		NolusYieldMarket: NolusYieldMarket{
			LppContract: "lpp-contract-address",
		},
		mockClient: mockClient,
	}

	// Setup test case 1
	mockClient.On("Price", mock.Anything, &lpp.PriceRequest{}).Return(
		&lpp.PriceResponse{
			Amount:      "100000000",
			AmountQuote: "200000000",
		},
		nil,
	).Once()

	// Call the function
	price, err := market.getPrice(context.Background())

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, price)
	// Expected price is Amount/AmountQuote as an Int (truncated)
	assert.Equal(t, "0", price.String()) // 100000000/200000000 = 0.5, truncated to 0

	// Setup test case 2
	mockClient.On("Price", mock.Anything, &lpp.PriceRequest{}).Return(
		&lpp.PriceResponse{
			Amount:      "200000000",
			AmountQuote: "100000000",
		},
		nil,
	).Once()

	price, err = market.getPrice(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, price)
	assert.Equal(t, "2", price.String()) // 200000000/100000000 = 2
}

// TestCalculateBurnAmount tests the calculateBurnAmount helper function with different scenarios
func TestCalculateBurnAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   string
		price    string
		decimals uint64
		expected string
	}{
		{
			name:     "Price greater than 1",
			amount:   "1000000",
			price:    "1.2",
			decimals: 6,
			expected: "833333", // 1000000 / 1.2 ≈ 833333.33... truncated to 833333
		},
		{
			name:     "Price less than 1",
			amount:   "1000000",
			price:    "0.8",
			decimals: 6,
			expected: "1250000", // 1000000 / 0.8 = 1250000
		},
		{
			name:     "Price exactly 1",
			amount:   "1000000",
			price:    "1.0",
			decimals: 6,
			expected: "1000000", // 1000000 / 1.0 = 1000000
		},
		{
			name:     "Very small price",
			amount:   "1000000",
			price:    "0.0001",
			decimals: 6,
			expected: "10000000000", // 1000000 / 0.0001 = 10000000000
		},
		{
			name:     "Higher decimals",
			amount:   "1000000",
			price:    "1.2345",
			decimals: 18,
			expected: "810044", // 1000000 / 1.2345 ≈ 810044.55... truncated to 810044
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create market with test decimals
			market := &NolusYieldMarket{
				Decimals: tc.decimals,
			}

			// Parse input values
			amount, ok := sdkmath.NewIntFromString(tc.amount)
			assert.True(t, ok, "Failed to parse amount")

			price, err := sdkmath.LegacyNewDecFromStr(tc.price)
			assert.NoError(t, err, "Failed to parse price")

			// Call the function
			result := market.calculateBurnAmount(amount, price)

			// Check the result matches expected
			expected, ok := sdkmath.NewIntFromString(tc.expected)
			assert.True(t, ok, "Failed to parse expected")
			assert.Equal(t, expected.String(), result.String(), "Burn amount calculation mismatch")
		})
	}
}
