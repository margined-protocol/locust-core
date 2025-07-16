package skipgo

import (
	"context"
	"flag"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

var integration = flag.Bool("integration", false, "run integration tests")

// TestPriceCurveAnalysis tests the price curve analysis approach
func TestPriceCurveAnalysis(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration test; use -integration flag to run")
	}

	logger := zaptest.NewLogger(t)

	// Initialize the Skip client
	skip, err := NewClient("https://api.skip.build")
	require.NoError(t, err, "Failed to initialize skip client")

	testCases := []struct {
		name           string
		tokenIn        string
		tokenOut       string
		amount         *big.Int
		maxPriceImpact float64
	}{
		{
			name:           "Small amount swap",
			tokenIn:        "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
			tokenOut:       "ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
			amount:         big.NewInt(10000000), // 10 ATOM
			maxPriceImpact: 1.0,                  // 1% max price impact
		},
		{
			name:           "Medium amount swap",
			tokenIn:        "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
			tokenOut:       "ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
			amount:         big.NewInt(100000000), // 100 ATOM
			maxPriceImpact: 3.0,                   // 3% max price impact
		},
		{
			name:           "Large amount swap",
			tokenIn:        "ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9",
			tokenOut:       "ibc/B559A80D62249C8AA07A380E2A2BEA6E5CA9A6F079C912C3A9E9B494105E4F81",
			amount:         big.NewInt(1000000000), // 1000 ATOM
			maxPriceImpact: 3.0,                    // 3% max price impact
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the findOptimalSwapAmount function
			optimalAmount, route, err := skip.FindOptimalSwapRoute(
				context.Background(),
				logger,
				"neutron-1",
				tc.tokenIn,
				tc.tokenOut,
				tc.amount,
				tc.maxPriceImpact,
			)

			// Verify the results
			require.NoError(t, err, "FindOptimalSwapRoute should not error")
			require.NotNil(t, optimalAmount, "Optimal amount should not be nil")
			require.NotNil(t, route, "Route should not be nil")

			// Verify amount constraints
			assert.LessOrEqual(t, optimalAmount.Cmp(tc.amount), 0,
				"Optimal amount should not exceed original amount")
			assert.Greater(t, optimalAmount.Cmp(big.NewInt(0)), 0,
				"Optimal amount should be greater than zero")

			// Log the results for debugging
			logger.Info("Swap test results",
				zap.String("test_case", tc.name),
				zap.String("original_amount", tc.amount.String()),
				zap.String("optimal_amount", optimalAmount.String()),
				zap.Any("route", route),
			)
		})
	}
}
