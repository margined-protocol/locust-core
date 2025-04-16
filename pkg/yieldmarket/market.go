package yieldmarket

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"go.uber.org/zap"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// YieldMarket defines an interface for any market that can provide yield
type YieldMarket interface {
	// GetChainID returns the chain ID for the market
	GetChainID() string

	// GetDenom returns the denom of the market
	GetDenom() string

	// GetCurrentRate returns the current interest/liquidity rate
	GetCurrentRate(ctx context.Context) (sdkmath.LegacyDec, error)

	// GetTotalLiquidity returns the total amount of underlying assets in the market
	GetTotalLiquidity(ctx context.Context) (sdkmath.Int, error)

	// GetTotalDebt returns the total amount of borrowed assets
	GetTotalDebt(ctx context.Context) (sdkmath.Int, error)

	// GetLentPosition returns the total amount lent including interest
	GetLentPosition(ctx context.Context) (sdkmath.Int, error)

	// CalculateRateWithUtilization simulates the rate after changing utilization
	CalculateRateWithUtilization(ctx context.Context, utilizationRate sdkmath.LegacyDec) (sdkmath.LegacyDec, error)

	// CalculateNewUtilization calculates the new utilization after adding/removing liquidity
	CalculateNewUtilization(ctx context.Context, liquidityChange sdkmath.Int, isDeposit bool) (sdkmath.LegacyDec, error)

	// MaximumWithdrawal returns the maximum amount that can be withdrawn
	MaximumWithdrawal(ctx context.Context) (sdkmath.Int, error)

	// LendFunds executes a deposit and lend
	LendFunds(ctx context.Context, amount sdkmath.Int) sdk.Msg

	// WithdrawFunds executes a repayment
	WithdrawFunds(ctx context.Context, amount sdkmath.Int) sdk.Msg

	// TransferFunds executes a transfer
	TransferFunds(ctx context.Context, source, destination, receiver string, amount sdkmath.Int) []sdk.Msg
}

// Retry function with exponential backoff
func retry(attempts int, sleep time.Duration, logger zap.Logger, fn func() error) error {
	for i := range make([]struct{}, attempts) {
		err := fn()
		if err == nil {
			return nil
		}

		// Log the attempt and error
		logger.Info("Attempt failed", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(sleep)

		// Exponential backoff
		sleep *= 2
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, "network error")
}
