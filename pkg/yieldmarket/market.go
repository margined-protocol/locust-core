package yieldmarket

import (
	"context"

	sdkmath "cosmossdk.io/math"

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
