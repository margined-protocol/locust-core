package perps

import (
	"testing"

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
		-9, // quantumConversionExponent
		-6, // atomicResolution
		response,
	)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, position)

	// Basic position fields
	require.Equal(t, "4890000000", position.EntryPrice.String(), "Entry price should be 4.89 * 10^9")
	require.Equal(t, "1000000", position.Amount.String(), "Size should be 1 * 10^6")
	require.Equal(t, "24566911", position.Margin.String(), "Margin should be 24.566911 * 10^6")
	// require.Equal(t, "0", position.CurrentPrice.String(), "Current Price should be 0")
	require.Equal(t, "-3255676", position.UnrealizedPnl.String(), "Unrealized PnL should be -0.003255676 * 10^9")
	require.Equal(t, "129000000", position.RealizedPnl.String(), "Realized PnL should be 0.129 * 10^9")
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
		"ATOM-USD",
		-9, // quantumConversionExponent
		-6, // atomicResolution
		response,
	)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, currentPrice)

	// Basic position fields
	require.Equal(t, "4890000000", currentPrice.String(), "Entry price should be 4.89 * 10^9")
}
