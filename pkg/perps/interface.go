package perps

import (
	"context"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

// Position represents a perpetual futures position
type Position struct {
	EntryPrice    sdkmath.Int
	Margin        sdkmath.Int
	Amount        sdkmath.Int
	CurrentPrice  sdkmath.Int
	UnrealizedPnl sdkmath.Int
	RealizedPnl   sdkmath.Int
}

// Provider defines the interface that any perps provider must implement
type Provider interface {
	// Initialization
	Initialize(ctx context.Context) error

	// Account Management
	CreateSubaccount(account string) (sdk.Msg, error)
	CheckSubaccount(account string) (bool, error)
	GetSubaccount() string
	GetAccountBalance() (sdk.Coins, error)
	GetSubaccountBalance() (sdk.Coins, error)

	// Order Management
	CreateMarketOrder(ctx context.Context, price, margin, size sdkmath.Int, isBuy, reduceOnly bool) ([]sdk.Msg, error)
	CreateLimitOrder(ctx context.Context, price, margin, size sdkmath.Int, isBuy, reduceOnly bool) ([]sdk.Msg, error)

	// Position Management
	GetPosition(ctx context.Context) (*Position, error)
	GetLiquidationPrice(equity, size, entryPrice, maintenanceMargin sdkmath.LegacyDec) sdkmath.LegacyDec

	// Fund Management
	DepositSubaccount(ctx context.Context, margin sdkmath.Int) ([]sdk.Msg, error)
	WithdrawSubaccount(ctx context.Context, margin sdkmath.Int) ([]sdk.Msg, error)

	// Provider Information
	GetProviderChainID() string
	GetProviderName() string
	GetProviderDenom() string
	GetProviderExecutor() string

	// Event Handling
	ProcessPerpEvent(events []abcitypes.Event) (currentPrice string, entryPrice string, err error)

	// High-Level Operations
	IncreasePosition(ctx context.Context, price float64, amount, margin sdkmath.Int, isLong bool) (*ExecutionResult, error)
	ReducePosition(ctx context.Context, price float64, amount, margin sdkmath.Int, isLong bool) (*ExecutionResult, error)
	ClosePosition(ctx context.Context, isLong bool) (*ExecutionResult, error)
	AdjustMargin(ctx context.Context, margin sdkmath.Int, isAdd bool) (*ExecutionResult, error)
}

// ExecutionResult represents the result of a position operation
type ExecutionResult struct {
	TxHash         string
	Events         []abcitypes.Event
	Position       *Position
	Executed       bool
	ExecutionPrice string
	Messages       []sdk.Msg // The messages that were or would be sent
	Notes          string    // Additional information about the execution
}
