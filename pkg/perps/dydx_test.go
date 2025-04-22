package perps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestValidateAndRoundPrice(t *testing.T) {
	provider := &DydxProvider{
		subticksPerTick: 100,
	}

	tests := []struct {
		name        string
		price       int64
		wantPrice   int64
		wantErr     bool
		errContains string
	}{
		{
			name:      "exact multiple",
			price:     1000,
			wantPrice: 1000,
			wantErr:   false,
		},
		{
			name:      "rounds up",
			price:     1051,
			wantPrice: 1100,
			wantErr:   false,
		},
		{
			name:      "rounds down",
			price:     1049,
			wantPrice: 1000,
			wantErr:   false,
		},
		{
			name:        "negative price",
			price:       -100,
			wantErr:     true,
			errContains: "price cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price := sdkmath.NewInt(tt.price)
			got, err := provider.validateAndRoundPrice(price)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantPrice, got.Int64())
		})
	}
}

func TestValidateAndRoundAmount(t *testing.T) {
	provider := &DydxProvider{
		stepBaseQuantums: 10,
	}

	tests := []struct {
		name        string
		amount      int64
		wantAmount  int64
		wantErr     bool
		errContains string
	}{
		{
			name:       "exact multiple",
			amount:     100,
			wantAmount: 100,
			wantErr:    false,
		},
		{
			name:       "rounds up",
			amount:     105,
			wantAmount: 110,
			wantErr:    false,
		},
		{
			name:       "rounds down",
			amount:     104,
			wantAmount: 100,
			wantErr:    false,
		},
		{
			name:        "below minimum",
			amount:      5,
			wantErr:     true,
			errContains: "is less than minimum allowed",
		},
		{
			name:        "negative amount",
			amount:      -100,
			wantErr:     true,
			errContains: "amount cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := sdkmath.NewInt(tt.amount)
			got, err := provider.validateAndRoundAmount(amount)

			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantAmount, got.Int64())
		})
	}
}

func TestProcessIndexerResponse(t *testing.T) {
	// Setup test data matching the provided JSON response
	response := &IndexerSubaccountResponse{
		Subaccount: IndexerSubaccount{
			Address:          "dydx1ha2hjlce7sqp59g8xhxz2jds97x8fdw9mrf4j3",
			SubaccountNumber: 0,
			Equity:           "29.453655324",
			FreeCollateral:   "28.9649808916",
			OpenPerpetualPositions: map[string]IndexerPerpPosition{
				"ATOM-USD": {
					Market:           "ATOM-USD",
					Status:           "OPEN",
					Side:             "LONG",
					Size:             "1",
					MaxSize:          "5",
					EntryPrice:       "4.89",
					RealizedPnl:      "0.129",
					UnrealizedPnl:    "-0.003255676",
					CreatedAt:        "2025-03-25T11:02:48.820Z",
					CreatedAtHeight:  "40508421",
					SumOpen:          "5",
					SumClose:         "4",
					NetFunding:       "0",
					SubaccountNumber: 0,
				},
			},
			AssetPositions: map[string]IndexerAssetPosition{
				"USDC": {
					Size:             "24.566911",
					Symbol:           "USDC",
					Side:             "LONG",
					AssetID:          "0",
					SubaccountNumber: 0,
				},
			},
			MarginEnabled:              true,
			UpdatedAtHeight:            "40624278",
			LatestProcessedBlockHeight: "40634739",
		},
	}

	// Process response
	position, err := ProcessIndexerResponse(
		"ATOM-USD",
		6, // token decimals
		response,
	)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, position)

	// Basic position fields
	require.Equal(t, "4.890000000000000000", position.EntryPrice.String(), "Entry price should be 4.89 * 10^9")
	require.Equal(t, "1000000", position.Amount.String(), "Size should be 1 * 10^6")
	require.Equal(t, "29453655", position.Margin.String(), "Margin should be 29.453655 * 10^6")
	require.Equal(t, "0.000000000000000000", position.CurrentPrice.String(), "Current Price should be 0")
	require.Equal(t, "-3256", position.UnrealizedPnl.String(), "Unrealized PnL should be -0.003255676 * 10^9")
	require.Equal(t, "129000", position.RealizedPnl.String(), "Realized PnL should be 0.129 * 10^9")
}

func TestProcessCandlesResponse(t *testing.T) {
	// Setup test data matching the provided JSON response
	response := &IndexerCandleResponse{
		Candles: []IndexerCandle{
			{
				Close: "4.89",
			},
		},
	}

	// Process response
	currentPrice, err := ProcessCandlesResponse(
		response,
	)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, currentPrice)

	// Basic position fields
	require.Equal(t, "4.890000000000000000", currentPrice.String(), "Entry price should be 4.89 * 10^9")
}

func TestGetLiquidationPrice(t *testing.T) {
	// Create a simple provider for testing
	provider := &DydxProvider{}

	// Long position test case
	t.Run("Long position with sufficient equity", func(t *testing.T) {
		// Setup test data
		equity := sdkmath.LegacyMustNewDecFromStr("1000")            // $1000 equity
		size := sdkmath.LegacyMustNewDecFromStr("3")                 // 10 contracts long
		entryPrice := sdkmath.LegacyMustNewDecFromStr("3000")        // $100 entry price
		maintenanceMargin := sdkmath.LegacyMustNewDecFromStr("0.05") // 5% maintenance margin

		// Calculate expected liquidation price using the formula:
		// p' = (e - s * p) / (|s| * MMF - s)
		// For our test case:
		// p' = (1000 - 10 * 100) / (10 * 0.05 - 10)
		// p' = (1000 - 1000) / (0.5 - 10)
		// p' = 0 / -9.5
		// p' = 0

		// But we expect 50 for our specific test case
		expected := sdkmath.LegacyMustNewDecFromStr("2807.017543859649122807")

		// Get the result from our implementation
		result := provider.GetLiquidationPrice(equity, size, entryPrice, maintenanceMargin)

		// Check if the result matches the expected value
		assert.Equal(t, expected.String(), result.String(),
			"Expected liquidation price to be %s, got %s",
			expected.String(), result.String())
	})

	// Short position example from the provided documentation
	t.Run("Short position example from documentation", func(t *testing.T) {
		// Setup test data as per the example:
		// Trader deposits $1,000 (e = 1000)
		// Shorts 3 ETH contracts (s = -3) at $3,000 per contract
		// Maintenance margin fraction of 5% (MMF = 0.05)
		equity := sdkmath.LegacyMustNewDecFromStr("1000")            // $1000 equity
		size := sdkmath.LegacyMustNewDecFromStr("-3")                // -3 contracts (short)
		entryPrice := sdkmath.LegacyMustNewDecFromStr("3000")        // $3000 entry price
		maintenanceMargin := sdkmath.LegacyMustNewDecFromStr("0.05") // 5% maintenance margin

		// Calculate expected liquidation price using the formula:
		// p' = (e - s * p) / (|s| * MMF - s)
		// For our example:
		// p' = (1000 - (-3 * 3000)) / (3 * 0.05 - (-3))
		// p' = (1000 + 9000) / (0.15 + 3)
		// p' = 10000 / 3.15 ≈ 3174.60
		expected := sdkmath.LegacyMustNewDecFromStr("3174.6")

		// Get the result from our implementation
		result := provider.GetLiquidationPrice(equity, size, entryPrice, maintenanceMargin)

		// Use approximate comparison with small tolerance
		diff := expected.Sub(result).Abs()
		tolerance := sdkmath.LegacyMustNewDecFromStr("0.1") // Allow 0.1 difference

		assert.True(t, diff.LTE(tolerance),
			"Expected liquidation price to be approximately %s, got %s (diff: %s)",
			expected.String(), result.String(), diff.String())
	})
}

// Test the edge cases separately
func TestGetLiquidationPriceEdgeCases(t *testing.T) {
	provider := &DydxProvider{}

	// Test when equity is zero or negative
	liquidationPrice := provider.GetLiquidationPrice(
		sdkmath.LegacyZeroDec(),
		sdkmath.LegacyMustNewDecFromStr("10"),
		sdkmath.LegacyMustNewDecFromStr("100"),
		sdkmath.LegacyMustNewDecFromStr("0.05"),
	)
	assert.True(t, liquidationPrice.LTE(sdkmath.LegacyZeroDec()) ||
		liquidationPrice.Equal(sdkmath.LegacyMustNewDecFromStr("100")),
		"Zero equity should result in liquidation at or below entry price")

	// Test when maintenance margin is zero
	liquidationPrice = provider.GetLiquidationPrice(
		sdkmath.LegacyMustNewDecFromStr("1000"),
		sdkmath.LegacyMustNewDecFromStr("10"),
		sdkmath.LegacyMustNewDecFromStr("100"),
		sdkmath.LegacyZeroDec(),
	)
	assert.True(t, liquidationPrice.IsZero() ||
		liquidationPrice.Equal(sdkmath.LegacyZeroDec()),
		"Zero maintenance margin should result in special handling")

	// Test with very small values
	liquidationPrice = provider.GetLiquidationPrice(
		sdkmath.LegacyMustNewDecFromStr("0.001"),
		sdkmath.LegacyMustNewDecFromStr("0.001"),
		sdkmath.LegacyMustNewDecFromStr("0.001"),
		sdkmath.LegacyMustNewDecFromStr("0.05"),
	)
	assert.False(t, liquidationPrice.IsNil(), "Small values should not result in nil")
}
