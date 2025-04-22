package yieldmarket

import (
	"context"
	"fmt"
	"time"

	conn "github.com/margined-protocol/locust-core/pkg/connection"
	"github.com/margined-protocol/locust-core/pkg/ibc"
	"github.com/margined-protocol/locust-core/pkg/math"
	// Import Umee leverage module types - you'll need to add these to your go.mod
	ltypes "github.com/margined-protocol/locust-core/pkg/proto/umee/leverage/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// UmeeYieldMarket implements the YieldMarket interface for Umee protocol
type UmeeYieldMarket struct {
	Name       string
	ChainID    string
	Prefix     string
	Denom      string
	Decimals   uint64
	Connection *grpc.ClientConn

	// For tracking state between calls
	cachedMarket *ltypes.QueryMarketSummaryResponse
	cachedToken  ltypes.Token
	lastUpdated  time.Time
	tokenUpdated time.Time

	// IBC Registry
	transferProvider ibc.TransferProvider

	// For transaction operations
	clientRegistry *conn.ClientRegistry
	signerAccount  string
	senderAddress  string

	logger *zap.Logger
}

// NewUmeeYieldMarket creates a new Umee market implementation
func NewUmeeYieldMarket(
	chainID string,
	prefix string,
	denom string,
	decimals uint64,
	connection *grpc.ClientConn,
	clientRegistry *conn.ClientRegistry,
	transferProvider ibc.TransferProvider,
	signerAccount string,
	senderAddress string,
	logger *zap.Logger,
) *UmeeYieldMarket {
	return &UmeeYieldMarket{
		ChainID:          chainID,
		Prefix:           prefix,
		Denom:            denom,
		Decimals:         decimals,
		Connection:       connection,
		transferProvider: transferProvider,
		clientRegistry:   clientRegistry,
		signerAccount:    signerAccount,
		senderAddress:    senderAddress,
		logger:           logger,
	}
}

// GetName returns the market identifier
func (u *UmeeYieldMarket) GetName() string {
	return u.Name
}

// GetChainID returns the chain ID for the market
func (u *UmeeYieldMarket) GetChainID() string {
	return u.ChainID
}

// GetDenom returns the denom of the price for the market
func (u *UmeeYieldMarket) GetDenom() string {
	return u.Denom
}

// refreshMarketData ensures we have up-to-date market data
func (u *UmeeYieldMarket) refreshMarketData(ctx context.Context) error {
	return retry(DefaultRetryAmount, 1*time.Second, *u.logger, func() error {
		// If data is less than 60 seconds old, don't refresh
		if u.cachedMarket != nil && time.Since(u.lastUpdated) < 60*time.Second {
			return nil
		}

		// Create Umee leverage query client
		queryClient := ltypes.NewQueryClient(u.Connection)

		// Fetch updated market data
		marketResp, err := queryClient.MarketSummary(ctx, &ltypes.QueryMarketSummary{
			Denom: u.Denom,
		})
		if err != nil {
			return fmt.Errorf("failed to fetch market data: %w", err)
		}

		u.cachedMarket = marketResp
		u.lastUpdated = time.Now()
		return nil
	})
}

// refreshTokenData ensures we have up-to-date token data
func (u *UmeeYieldMarket) refreshTokenData(ctx context.Context) error {
	// If data exists and is less than 1 hour old, don't refresh
	if !u.tokenUpdated.IsZero() && time.Since(u.tokenUpdated) < 60*60*time.Second {
		return nil
	}

	// Create Umee leverage query client
	queryClient := ltypes.NewQueryClient(u.Connection)

	// Fetch all registered tokens
	tokensResp, err := queryClient.RegisteredTokens(ctx, &ltypes.QueryRegisteredTokens{})
	if err != nil {
		return fmt.Errorf("failed to fetch registered tokens: %w", err)
	}

	// Find the token matching our denom
	for _, token := range tokensResp.Registry {
		if token.BaseDenom == u.Denom {
			u.cachedToken = token // This should work since Registry is a slice of Token (not pointers)
			u.tokenUpdated = time.Now()
			return nil
		}
	}

	return fmt.Errorf("token with denom %s not found in registered tokens", u.Denom)
}

// GetCurrentRate returns the current interest/liquidity rate
func (u *UmeeYieldMarket) GetCurrentRate(ctx context.Context) (sdkmath.LegacyDec, error) {
	if err := u.refreshMarketData(ctx); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	// In Umee, we use the SupplyAPY as the liquidity rate
	return u.cachedMarket.Supply_APY, nil
}

// GetTotalLiquidity returns the total amount of underlying assets
func (u *UmeeYieldMarket) GetTotalLiquidity(ctx context.Context) (sdkmath.Int, error) {
	if err := u.refreshMarketData(ctx); err != nil {
		return sdkmath.Int{}, err
	}

	// Total supplied is uToken supply * exchange rate
	return u.cachedMarket.Supplied, nil
}

// GetTotalDebt returns the total amount of borrowed assets
func (u *UmeeYieldMarket) GetTotalDebt(ctx context.Context) (sdkmath.Int, error) {
	if err := u.refreshMarketData(ctx); err != nil {
		return sdkmath.Int{}, err
	}

	// Total debt is the borrowed amount
	return u.cachedMarket.Borrowed, nil
}

// GetLentPosition returns the total amount lent including interest
func (u *UmeeYieldMarket) GetLentPosition(ctx context.Context) (sdkmath.Int, error) {
	lentAmount := sdkmath.ZeroInt()
	err := retry(DefaultRetryAmount, 1*time.Second, *u.logger, func() error {
		// Create Umee leverage query client
		queryClient := ltypes.NewQueryClient(u.Connection)

		// Fetch account positions
		positionResp, err := queryClient.AccountBalances(ctx, &ltypes.QueryAccountBalances{
			Address: u.senderAddress,
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// No positions found for account
				return nil
			}
			return fmt.Errorf("failed to fetch account positions: %w", err)
		}

		// Find the collateral position for this denom
		for _, collateral := range positionResp.Supplied {
			if collateral.Denom == u.Denom {
				// Return the collateral amount
				lentAmount = collateral.Amount
				return nil
			}
		}

		// No position found for this denom
		return nil
	})
	if err != nil {
		return sdkmath.ZeroInt(), err
	}

	return lentAmount, nil
}

// deriveBorrowAPY calculates the borrow rate based on utilization and token parameters
// following Umee's interest rate model
func (u *UmeeYieldMarket) deriveBorrowAPY(token *ltypes.Token, utilization sdkmath.LegacyDec) sdkmath.LegacyDec {
	// Tokens which have reached or exceeded their max supply utilization always use max borrow APY
	if utilization.GTE(token.MaxSupplyUtilization) {
		return token.MaxBorrowRate
	}

	// Tokens which are past kink value but have not reached max supply utilization interpolate between the two
	if utilization.GTE(token.KinkUtilization) {
		return math.Interpolate(
			utilization,                // x
			token.KinkUtilization,      // x1
			token.KinkBorrowRate,       // y1
			token.MaxSupplyUtilization, // x2
			token.MaxBorrowRate,        // y2
		)
	}

	// utilization is between 0% and kink value
	return math.Interpolate(
		utilization,             // x
		sdkmath.LegacyZeroDec(), // x1
		token.BaseBorrowRate,    // y1
		token.KinkUtilization,   // x2
		token.KinkBorrowRate,    // y2
	)
}

// deriveSupplyAPY calculates the supply rate based on borrow rate, utilization and reduction factors
func (u *UmeeYieldMarket) deriveSupplyAPY(
	borrowRate, utilization, oracleRewardFactor, rewardsAuctionFee, reserveFactor sdkmath.LegacyDec,
) sdkmath.LegacyDec {
	// Compile the reduction from all factors
	reduction := oracleRewardFactor.Add(rewardsAuctionFee).Add(reserveFactor)

	// supply APY = borrow APY * utilization * (1 - reduction)
	return borrowRate.Mul(utilization).Mul(sdkmath.LegacyOneDec().Sub(reduction))
}

// CalculateRateWithUtilization simulates the rate after changing utilization
func (u *UmeeYieldMarket) CalculateRateWithUtilization(ctx context.Context, utilizationRate sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	// Refresh market data
	if err := u.refreshMarketData(ctx); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	// Refresh token data
	if err := u.refreshTokenData(ctx); err != nil {
		// If we can't get token data, fall back to the approximation method
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to fetch token data: %w", err)
	}

	// Get the module parameters for reduction factors
	queryClient := ltypes.NewQueryClient(u.Connection)
	paramsResp, err := queryClient.Params(ctx, &ltypes.QueryParams{})
	if err != nil {
		// If we can't get the params, fall back to approximation
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to fetch params: %w", err)
	}

	// Calculate borrow rate using the token parameters
	borrowRate := u.deriveBorrowAPY(&u.cachedToken, utilizationRate)

	// Calculate supply rate using the derived borrow rate and reduction factors
	supplyRate := u.deriveSupplyAPY(
		borrowRate,
		utilizationRate,
		paramsResp.Params.OracleRewardFactor,
		paramsResp.Params.RewardsAuctionFee,
		u.cachedToken.ReserveFactor,
	)

	return supplyRate, nil
}

// CalculateNewUtilization calculates the new utilization after adding/removing liquidity
func (u *UmeeYieldMarket) CalculateNewUtilization(ctx context.Context, liquidityChange sdkmath.Int, isDeposit bool) (sdkmath.LegacyDec, error) {
	debt, err := u.GetTotalDebt(ctx)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	liquidity, err := u.GetTotalLiquidity(ctx)
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

	// Calculate new utilization rate
	utilizationRate := sdkmath.LegacyNewDecFromInt(debt).Quo(sdkmath.LegacyNewDecFromInt(newLiquidity))

	// Cap at 100%
	one := sdkmath.LegacyOneDec()
	if utilizationRate.GT(one) {
		utilizationRate = one
	}

	return utilizationRate, nil
}

// MaximumWithdrawal returns the maximum amount that can be withdrawn
func (u *UmeeYieldMarket) MaximumWithdrawal(ctx context.Context) (sdkmath.Int, error) {
	// Create Umee leverage query client
	queryClient := ltypes.NewQueryClient(u.Connection)

	// Get maximum withdrawable amount
	withdrawResp, err := queryClient.MaxWithdraw(ctx, &ltypes.QueryMaxWithdraw{
		Address: u.senderAddress,
		Denom:   u.Denom,
	})
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to fetch max withdraw: %w", err)
	}
	// Find the collateral position for this denom
	for _, collateral := range withdrawResp.Tokens {
		if collateral.Denom == u.Denom {
			return collateral.Amount, nil
		}
	}

	// No position found for this denom
	return sdkmath.ZeroInt(), nil
}

// LendFunds deposits funds into Umee leverage module
func (u *UmeeYieldMarket) LendFunds(_ context.Context, amount sdkmath.Int) sdk.Msg {
	// Create MsgSupply message
	supplyMsg := ltypes.MsgSupply{
		Supplier: u.senderAddress,
		Asset:    sdk.NewCoin(u.Denom, amount),
	}

	return &supplyMsg
}

// WithdrawFunds withdraws funds from Umee leverage module
func (u *UmeeYieldMarket) WithdrawFunds(ctx context.Context, amount sdkmath.Int) sdk.Msg {
	if err := u.refreshMarketData(ctx); err != nil {
		return nil
	}

	uTokenExchangeRate, err := u.cachedMarket.UTokenExchangeRate.Float64()
	if err != nil {
		return nil
	}

	// Convert the amount to uToken
	utokenAmount := math.DivideWithDecimals(uTokenExchangeRate, amount.BigInt(), int(u.cachedMarket.Exponent))

	u.logger.Debug("📊 Withdrawing from Umee",
		zap.String("amount", amount.String()),
		zap.String("utoken_exchange_rate", fmt.Sprintf("%f", uTokenExchangeRate)),
		zap.String("utoken_amount", utokenAmount.String()),
	)

	// Create MsgWithdraw message
	withdrawMsg := ltypes.MsgWithdraw{
		Supplier: u.senderAddress,
		// Umee requires the asset to be prefixed with "u/" as you are withdrawing uToken
		Asset: sdk.NewCoin("u/"+u.Denom, sdkmath.NewIntFromBigInt(utokenAmount)),
	}

	return &withdrawMsg
}

// TransferFunds executes a transfer between markets
func (u *UmeeYieldMarket) TransferFunds(ctx context.Context, source, destination, receiver string, amount sdkmath.Int) []sdk.Msg {
	// First withdraw from Umee
	withdrawMsg := u.WithdrawFunds(ctx, amount)
	if withdrawMsg == nil {
		return nil
	}

	// Setup coin to transfer
	coin := sdk.NewCoin(u.Denom, amount)

	transferMsg, err := u.transferProvider.CreateTransferMsg(ctx, &ibc.TransferRequest{
		SourceChain:      source,
		DestinationChain: destination,
		Amount:           coin,
		Timeout:          10,
		Sender:           u.senderAddress,
		Receiver:         receiver,
	})
	if err != nil {
		return nil
	}

	return []sdk.Msg{withdrawMsg, transferMsg}
}
