package yieldmarket

import (
	"context"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	conn "github.com/margined-protocol/locust-core/pkg/connection"
	"github.com/margined-protocol/locust-core/pkg/contracts/nolus/lpp"
	"github.com/margined-protocol/locust-core/pkg/ibc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NolusYieldMarket implements the YieldMarket interface for Nolus protocol
type NolusYieldMarket struct {
	Name        string
	ChainID     string
	Prefix      string
	Denom       string
	Decimals    uint64
	LppContract string

	Connection *grpc.ClientConn

	// For tracking state between calls
	cachedLppBalance *lpp.LppBalanceResponse
	cachedPrice      *lpp.PriceResponse
	lastUpdated      time.Time

	// IBC Registry
	transferProvider ibc.TransferProvider

	// For transaction operations
	clientRegistry *conn.ClientRegistry
	signerAccount  string
	senderAddress  string

	logger *zap.Logger
}

// NewNolusYieldMarket creates a new Nolus market implementation
func NewNolusYieldMarket(
	chainID string,
	prefix string,
	denom string,
	decimals uint64,
	lppContract string,
	connection *grpc.ClientConn,
	clientRegistry *conn.ClientRegistry,
	transferProvider ibc.TransferProvider,
	signerAccount string,
	senderAddress string,
	logger *zap.Logger,
) *NolusYieldMarket {
	return &NolusYieldMarket{
		ChainID:          chainID,
		Prefix:           prefix,
		Denom:            denom,
		Decimals:         decimals,
		LppContract:      lppContract,
		Connection:       connection,
		transferProvider: transferProvider,
		clientRegistry:   clientRegistry,
		signerAccount:    signerAccount,
		senderAddress:    senderAddress,
	}
}

// GetName returns the market identifier
func (n *NolusYieldMarket) GetName() string {
	return n.Name
}

// GetChainID returns the chain ID for the market
func (n *NolusYieldMarket) GetChainID() string {
	return n.ChainID
}

// GetDenom returns the denom of the price for the market
func (n *NolusYieldMarket) GetDenom() string {
	return n.Denom
}

// refreshMarketData ensures we have up-to-date market data
func (n *NolusYieldMarket) refreshMarketData(ctx context.Context) error {
	return retry(5, 1*time.Second, *n.logger, func() error {
		// If data is less than 60 seconds old, don't refresh
		if n.cachedLppBalance != nil && time.Since(n.lastUpdated) < 60*time.Second {
			return nil
		}

		// Fetch updated LPP balance
		lppClient := lpp.NewQueryClient(n.Connection, n.LppContract)

		// Get LppBalance
		balance, err := lppClient.LppBalance(
			ctx,
			&lpp.LppBalanceRequest{},
		)
		if err != nil {
			return fmt.Errorf("failed to fetch LPP balance: %w", err)
		}

		// Get current price
		price, err := lppClient.Price(
			ctx,
			&lpp.PriceRequest{},
		)
		if err != nil {
			return fmt.Errorf("failed to fetch price: %w", err)
		}

		n.cachedLppBalance = balance
		n.cachedPrice = price
		n.lastUpdated = time.Now()
		return nil
	})
}

// getPrice ensures we have up-to-date market data
func (n *NolusYieldMarket) getPrice(ctx context.Context) (sdkmath.LegacyDec, error) {
	lppClient := lpp.NewQueryClient(n.Connection, n.LppContract)

	// Get current price
	res, err := lppClient.Price(
		ctx,
		&lpp.PriceRequest{},
	)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to fetch price: %w", err)
	}

	amount, err := sdkmath.LegacyNewDecFromStr(res.Amount)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to parse amount: %w", err)
	}

	amountQuote, err := sdkmath.LegacyNewDecFromStr(res.AmountQuote)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to parse amount quote: %w", err)
	}

	price := amount.Quo(amountQuote)

	return price, nil
}

// GetCurrentRate returns the current interest/liquidity rate
func (n *NolusYieldMarket) GetCurrentRate(ctx context.Context) (sdkmath.LegacyDec, error) {
	if err := n.refreshMarketData(ctx); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	// In Nolus, we need to derive the rate from the price changes
	// This is a placeholder - actual implementation would need to calculate
	// APR from current price information

	return sdkmath.LegacyNewDec(0), fmt.Errorf("not implemented: rate calculation needs price history")
}

// GetTotalLiquidity returns the total amount of underlying assets
func (n *NolusYieldMarket) GetTotalLiquidity(ctx context.Context) (sdkmath.Int, error) {
	if err := n.refreshMarketData(ctx); err != nil {
		return sdkmath.Int{}, err
	}

	// Extract total liquidity from cached balance
	// Assuming the balance has a field for total LPP balance
	liquidity, success := sdkmath.NewIntFromString(n.cachedLppBalance.Balance.Amount)
	if !success {
		return sdkmath.Int{}, fmt.Errorf("invalid balance amount")
	}

	return liquidity, nil
}

// GetTotalDebt returns the total amount of borrowed assets
func (n *NolusYieldMarket) GetTotalDebt(ctx context.Context) (sdkmath.Int, error) {
	if err := n.refreshMarketData(ctx); err != nil {
		return sdkmath.Int{}, err
	}

	// Extract total debt (principal + interest due)
	principalDue, success := sdkmath.NewIntFromString(n.cachedLppBalance.TotalPrincipalDue.Amount)
	if !success {
		return sdkmath.Int{}, fmt.Errorf("invalid principal due")
	}

	interestDue, success := sdkmath.NewIntFromString(n.cachedLppBalance.TotalInterestDue.Amount)
	if !success {
		return sdkmath.Int{}, fmt.Errorf("invalid interest due")
	}

	return principalDue.Add(interestDue), nil
}

// GetLentPosition returns the total amount lent including interest
func (n *NolusYieldMarket) GetLentPosition(ctx context.Context) (sdkmath.Int, error) {
	lentAmount := sdkmath.ZeroInt()
	err := retry(5, 1*time.Second, *n.logger, func() error {
		// Query the LPP contract for the user's balance
		lppClient := lpp.NewQueryClient(n.Connection, n.LppContract)

		balanceResp, err := lppClient.Balance(
			ctx,
			&lpp.BalanceRequest{
				Address: n.senderAddress,
			},
		)
		if err != nil {
			return nil
		}

		lentAmount = sdkmath.NewIntFromUint64(balanceResp.Balance.Uint64())
		return nil
	})

	if err != nil {
		return sdkmath.ZeroInt(), err
	}

	return lentAmount, nil
}

// CalculateRateWithUtilization simulates the rate after changing utilization
func (n *NolusYieldMarket) CalculateRateWithUtilization(ctx context.Context, utilizationRate sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	// Nolus might not have a direct API for this, we would need to implement
	// based on the interest rate model from Nolus
	return sdkmath.LegacyNewDec(0), fmt.Errorf("not implemented: Nolus interest rate simulation")
}

// CalculateNewUtilization calculates the new utilization after adding/removing liquidity
func (n *NolusYieldMarket) CalculateNewUtilization(ctx context.Context, liquidityChange sdkmath.Int, isDeposit bool) (sdkmath.LegacyDec, error) {
	debt, err := n.GetTotalDebt(ctx)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	liquidity, err := n.GetTotalLiquidity(ctx)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	// Calculate new liquidity after change
	newLiquidity := liquidity
	if isDeposit {
		newLiquidity = newLiquidity.Add(liquidityChange)
	} else {
		// Ensure we don't withdraw more than available
		if liquidityChange.GT(liquidity) {
			return sdkmath.LegacyDec{}, fmt.Errorf("cannot withdraw more than available liquidity")
		}
		newLiquidity = newLiquidity.Sub(liquidityChange)
	}

	// If new liquidity is zero, utilization is 100%
	if newLiquidity.IsZero() {
		return sdkmath.LegacyOneDec(), nil
	}

	debtDec := sdkmath.LegacyNewDecFromInt(debt)
	totalLiquidityDec := sdkmath.LegacyNewDecFromInt(newLiquidity).Add(debtDec)

	// Calculate new utilization rate
	utilizationRate := debtDec.Quo(totalLiquidityDec)

	// Cap at 100%
	one := sdkmath.LegacyOneDec()
	if utilizationRate.GT(one) {
		utilizationRate = one
	}

	return utilizationRate, nil
}

// MaximumWithdrawal returns the maximum amount that can be withdrawn
func (n *NolusYieldMarket) MaximumWithdrawal(ctx context.Context) (sdkmath.Int, error) {
	// Get the user's balance in the LPP
	lppClient := lpp.NewQueryClient(n.Connection, n.LppContract)

	balanceResp, err := lppClient.Balance(
		ctx,
		&lpp.BalanceRequest{
			Address: n.senderAddress,
		},
	)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to fetch balance: %w", err)
	}

	// Get deposit capacity to check if withdrawal is constrained by utilization
	capacityResp, err := lppClient.DepositCapacity(
		ctx,
		&lpp.DepositCapacityRequest{},
	)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to fetch deposit capacity: %w", err)
	}

	// If deposit capacity exists, it means the pool has utilization constraints
	if capacityResp.Capacity != nil {
		// Implementation would need to calculate maximum withdrawal considering utilization
		// This is a placeholder
	}

	// Return the balance as maximum withdrawal
	return sdkmath.NewIntFromUint64(balanceResp.Balance.Uint64()), nil
}

// LendFunds deposits funds into the LPP
func (n *NolusYieldMarket) LendFunds(ctx context.Context, amount sdkmath.Int) sdk.Msg {
	// Create deposit message for the LPP contract
	depositMsg, err := lpp.BuildDepositMsg(n.senderAddress, n.LppContract, sdk.NewCoins(sdk.NewCoin(n.Denom, amount)))
	if err != nil {
		return nil
	}

	return depositMsg
}

// calculateBurnAmount converts a withdrawal amount to the corresponding burn amount using price
// It properly handles decimal scaling to prevent precision loss
func (n *NolusYieldMarket) calculateBurnAmount(amount sdkmath.Int, price sdkmath.LegacyDec) sdkmath.Int {
	// Calculate the inverse price (1/price)
	invPrice := sdkmath.LegacyOneDec().Quo(price)

	// Scale the inverse price by 10^decimals to preserve precision when converting to Int
	scaleFactor := sdkmath.LegacyNewDec(10).Power(uint64(n.Decimals))
	scaledInvPrice := invPrice.Mul(scaleFactor)

	// Convert to Int preserving decimals, then multiply by amount
	scaledAmount := amount.Mul(scaledInvPrice.TruncateInt())

	// Divide by scale factor to get the final amount in the correct units
	// Calculate 10^decimals for the divisor
	divisor := sdkmath.NewInt(1)
	for i := uint64(0); i < n.Decimals; i++ {
		divisor = divisor.MulRaw(10)
	}
	return scaledAmount.Quo(divisor)
}

// WithdrawFunds withdraws funds from the LPP
func (n *NolusYieldMarket) WithdrawFunds(ctx context.Context, amount sdkmath.Int) sdk.Msg {
	price, err := n.getPrice(ctx)
	if err != nil {
		return nil
	}

	// Calculate the amount to burn using our helper function
	amountToBurn := n.calculateBurnAmount(amount, price)

	// Create burn message for the LPP contract
	burnMsg, err := lpp.BuildBurnMsg(n.senderAddress, n.LppContract, amountToBurn.Uint64())
	if err != nil {
		return nil
	}

	return burnMsg
}

// TransferFunds executes a transfer between markets
func (n *NolusYieldMarket) TransferFunds(ctx context.Context, source, destination, receiver string, amount sdkmath.Int) []sdk.Msg {
	// First withdraw from LPP
	withdrawMsg := n.WithdrawFunds(ctx, amount)
	if withdrawMsg == nil {
		return nil
	}

	// Setup coin to transfer
	coin := sdk.NewCoin(n.Denom, amount)

	transferMsg, err := n.transferProvider.CreateTransferMsg(ctx, &ibc.TransferRequest{
		SourceChain:      source,
		DestinationChain: destination,
		Amount:           coin,
		Timeout:          10,
		Sender:           n.senderAddress,
		Receiver:         receiver,
	})
	if err != nil {
		return nil
	}

	return []sdk.Msg{withdrawMsg, transferMsg}
}
