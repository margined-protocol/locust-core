package yieldmarket

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/assert"

	ltypes "github.com/margined-protocol/locust-core/pkg/proto/umee/leverage/types"
)

// TestDeriveBorrowAPY tests the borrow rate calculation function
func TestDeriveBorrowAPY(t *testing.T) {
	// Create test token with Umee's standard parameters
	testToken := ltypes.Token{
		BaseDenom:            "uumee",
		SymbolDenom:          "UMEE",
		Exponent:             6,
		ReserveFactor:        sdkmath.LegacyNewDecWithPrec(20, 2),  // 0.20 or 20%
		CollateralWeight:     sdkmath.LegacyNewDecWithPrec(80, 2),  // 0.80 or 80%
		LiquidationThreshold: sdkmath.LegacyNewDecWithPrec(85, 2),  // 0.85 or 85%
		BaseBorrowRate:       sdkmath.LegacyNewDecWithPrec(2, 2),   // 0.02 or 2%
		KinkUtilization:      sdkmath.LegacyNewDecWithPrec(80, 2),  // 0.80 or 80%
		KinkBorrowRate:       sdkmath.LegacyNewDecWithPrec(22, 2),  // 0.22 or 22%
		MaxSupplyUtilization: sdkmath.LegacyNewDecWithPrec(90, 2),  // 0.90 or 90%
		MaxBorrowRate:        sdkmath.LegacyNewDecWithPrec(152, 2), // 1.52 or 152%
		Blacklist:            false,
	}

	// Create test cases
	testCases := []struct {
		name        string
		utilization string
		expected    string
	}{
		{
			name:        "Zero utilization",
			utilization: "0.0",
			expected:    "0.02", // Base borrow rate
		},
		{
			name:        "Mid utilization (40%)",
			utilization: "0.4",
			expected:    "0.12", // Linear interpolation from base to kink
		},
		{
			name:        "Kink point (80%)",
			utilization: "0.8",
			expected:    "0.22", // Exactly at kink point
		},
		{
			name:        "Above kink (85%)",
			utilization: "0.85",
			expected:    "0.87", // Linear interpolation from kink to max
		},
		{
			name:        "Max utilization (90%)",
			utilization: "0.9",
			expected:    "1.52", // Max borrow rate
		},
		{
			name:        "Above max utilization (95%)",
			utilization: "0.95",
			expected:    "1.52", // Still capped at max borrow rate
		},
	}

	// Create a market for testing
	market := UmeeYieldMarket{}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			utilization, _ := sdkmath.LegacyNewDecFromStr(tc.utilization)
			expected, _ := sdkmath.LegacyNewDecFromStr(tc.expected)

			// Call the function
			result := market.deriveBorrowAPY(&testToken, utilization)

			// Check result
			delta := sdkmath.LegacyNewDecWithPrec(1, 6) // 0.000001 tolerance
			assert.InDelta(t,
				expected.MustFloat64(),
				result.MustFloat64(),
				delta.MustFloat64(),
				"Borrow APY at %s utilization should be %s, got %s",
				tc.utilization, tc.expected, result.String())
		})
	}
}
