package evaluator

import (
	"math"
	"testing"

	"github.com/margined-protocol/locust-core/pkg/contracts/levana/market"
)

// ComputeFundingRates function as implemented

func TestComputeFundingRates(t *testing.T) {
	tests := []struct {
		name                          string
		longNotional                  float64
		shortNotional                 float64
		fundingRateSensitivity        float64
		fundingRateMaxAnnualized      float64
		deltaNeutralityFeeSensitivity float64
		deltaNeutralityFeeCap         float64
		expectedLongRate              float64
		expectedShortRate             float64
	}{
		{
			name:                          "Longs more popular",
			longNotional:                  238233.573628609301302489,
			shortNotional:                 195098.216241976218375226,
			fundingRateSensitivity:        2.0,
			fundingRateMaxAnnualized:      0.9,
			deltaNeutralityFeeCap:         0.005,
			deltaNeutralityFeeSensitivity: 100000000,
			expectedLongRate:              0.199086973976755554,
			expectedShortRate:             -0.243104228152260379,
		},
		{
			name:                          "Shorts more popular",
			longNotional:                  1770.597499123069530388,
			shortNotional:                 2684.059794083737358887,
			fundingRateSensitivity:        1.5,
			fundingRateMaxAnnualized:      0.9,
			deltaNeutralityFeeCap:         0.0002,
			deltaNeutralityFeeSensitivity: 17006505,
			expectedLongRate:              -0.466272633355340797,
			expectedShortRate:             0.307586723793656979,
		},
		{
			name:                          "No liquidity both sides",
			longNotional:                  0,
			shortNotional:                 0,
			fundingRateSensitivity:        1.5,
			fundingRateMaxAnnualized:      0.9,
			deltaNeutralityFeeCap:         0.0002,
			deltaNeutralityFeeSensitivity: 18962128239885,
			expectedLongRate:              0,
			expectedShortRate:             0,
		},
		{
			name:                          "wBTC failure case",
			longNotional:                  62533.63301,
			shortNotional:                 38504.06259,
			fundingRateSensitivity:        1,
			fundingRateMaxAnnualized:      0.45,
			deltaNeutralityFeeCap:         0.0002,
			deltaNeutralityFeeSensitivity: 50000000000,
			expectedLongRate:              0.2378277759,
			expectedShortRate:             -0.3862510566,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			longRate, shortRate := ComputeFundingRates(
				tt.longNotional,
				tt.shortNotional,
				tt.fundingRateSensitivity,
				tt.fundingRateMaxAnnualized,
				tt.deltaNeutralityFeeSensitivity,
				tt.deltaNeutralityFeeCap,
			)

			if math.Abs(longRate-tt.expectedLongRate) > 1e-6 {
				t.Errorf("ComputeFundingRates() longRate = %v, expected %v", longRate, tt.expectedLongRate)
			}
			if math.Abs(shortRate-tt.expectedShortRate) > 1e-6 {
				t.Errorf("ComputeFundingRates() shortRate = %v, expected %v", shortRate, tt.expectedShortRate)
			}
		})
	}
}

// TestCheckPositionExit tests the CheckPositionExit method.
func TestCheckPositionExit(t *testing.T) {
	tests := []struct {
		name               string
		pos                market.Position
		currentPrice       float64
		currentRate        float64
		expectedShouldExit bool
		expectedReason     string
	}{
		{
			name: "Long near liquidation",
			pos: market.Position{
				ID:                   "1",
				DirectionToBase:      "long",
				LiquidationPriceBase: "90",  // Example liquidation price
				TakeProfitPriceBase:  "150", // Example take profit price
			},
			// With a 5% buffer, long position should exit when currentPrice is ≤ 90 * 1.05 = 94.5.
			currentPrice:       94.5,
			currentRate:        -0.1, // negative funding rate
			expectedShouldExit: true,
			expectedReason:     "liquidation-risk",
		},
		{
			name: "Long near take profit",
			pos: market.Position{
				ID:                   "2",
				DirectionToBase:      "long",
				LiquidationPriceBase: "80",
				TakeProfitPriceBase:  "120",
			},
			// For long positions, exit if currentPrice is near the take profit target.
			// With a 5% buffer, currentPrice should be ≥ 120 * (1 - 0.05) = 114.
			currentPrice:       114.0,
			currentRate:        -0.1,
			expectedShouldExit: true,
			expectedReason:     "take-profit",
		},
		{
			name: "Long hold",
			pos: market.Position{
				ID:                   "3",
				DirectionToBase:      "long",
				LiquidationPriceBase: "80",
				TakeProfitPriceBase:  "150",
			},
			// Current price is comfortably above liquidation and below take profit.
			currentPrice:       100.0,
			currentRate:        -0.1,
			expectedShouldExit: false,
			expectedReason:     "",
		},
		{
			name: "Short near liquidation",
			pos: market.Position{
				ID:                   "4",
				DirectionToBase:      "short",
				LiquidationPriceBase: "120", // For short positions, liquidation is above currentPrice.
				TakeProfitPriceBase:  "80",
			},
			// For shorts, with a 5% buffer, currentPrice should be ≥ 120 * 0.95 = 114.
			currentPrice:       114.0,
			currentRate:        -0.1,
			expectedShouldExit: true,
			expectedReason:     "liquidation-risk",
		},
		{
			name: "Short near take profit",
			pos: market.Position{
				ID:                   "5",
				DirectionToBase:      "short",
				LiquidationPriceBase: "150",
				TakeProfitPriceBase:  "100", // For shorts, take profit is below currentPrice.
			},
			// With a 5% buffer, currentPrice should be ≤ 100 * 1.05 = 105.
			currentPrice:       105.0,
			currentRate:        -0.1,
			expectedShouldExit: true,
			expectedReason:     "take-profit",
		},
		{
			name: "Short hold",
			pos: market.Position{
				ID:                   "6",
				DirectionToBase:      "short",
				LiquidationPriceBase: "150",
				TakeProfitPriceBase:  "100",
			},
			// Current price is comfortably away from both targets.
			currentPrice:       130.0,
			currentRate:        -0.1,
			expectedShouldExit: false,
			expectedReason:     "",
		},
		{
			name: "Funding positive for long",
			pos: market.Position{
				ID:                   "7",
				DirectionToBase:      "long",
				LiquidationPriceBase: "80",
				TakeProfitPriceBase:  "150",
			},
			// Here the funding rate is positive, which should trigger an exit.
			currentPrice:       100.0,
			currentRate:        0.1,
			expectedShouldExit: true,
			expectedReason:     "funding-positive",
		},
		{
			name: "Funding positive for short",
			pos: market.Position{
				ID:                   "8",
				DirectionToBase:      "short",
				LiquidationPriceBase: "150",
				TakeProfitPriceBase:  "100",
			},
			currentPrice:       130.0,
			currentRate:        0.1,
			expectedShouldExit: true,
			expectedReason:     "funding-positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the revised CheckPositionExit with currentRate and currentPrice.
			shouldExit, reason := (&FundingRateEvaluator{}).CheckPositionExit(tt.pos, tt.currentRate, tt.currentPrice)
			if shouldExit != tt.expectedShouldExit {
				t.Errorf("CheckPositionExit() for %s returned shouldExit=%v; expected %v", tt.name, shouldExit, tt.expectedShouldExit)
			}
			if reason != tt.expectedReason {
				t.Errorf("CheckPositionExit() for %s returned reason=%q; expected %q", tt.name, reason, tt.expectedReason)
			}
		})
	}
}
