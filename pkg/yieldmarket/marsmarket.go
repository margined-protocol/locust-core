package yieldmarket

import (
	"context"
	"fmt"
	"time"

	conn "github.com/margined-protocol/locust-core/pkg/connection"
	cm "github.com/margined-protocol/locust-core/pkg/contracts/mars/creditmanager"
	rb "github.com/margined-protocol/locust-core/pkg/contracts/mars/redbank"
	"github.com/margined-protocol/locust-core/pkg/ibc"
	"google.golang.org/grpc"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MarsYieldMarket implements the YieldMarket interface for Mars protocol
type MarsYieldMarket struct {
	Name          string
	ChainID       string
	Prefix        string
	OracleDenom   string
	CreditAccount uint64
	CreditManager string
	Redbank       string

	Connection *grpc.ClientConn

	// For tracking state between calls
	cachedMarket *rb.MarketV2Response
	lastUpdated  time.Time

	// IBC Registry
	IBCRegistry *ibc.ConnectionRegistry

	// Added for TransferFunds method
	clientRegistry *conn.ClientRegistry
	signerAccount  string
	senderAddress  string
}

// NewMarsYieldMarket creates a new Mars market implementation
func NewMarsYieldMarket(
	chainID string,
	prefix string,
	oracleDenom string,
	redbank string,
	creditManager string,
	creditAccount uint64,
	connection *grpc.ClientConn,
	clientRegistry *conn.ClientRegistry,
	ibcRegistry *ibc.ConnectionRegistry,
	signerAccount string,
	senderAddress string,
) *MarsYieldMarket {
	return &MarsYieldMarket{
		ChainID:        chainID,
		Prefix:         prefix,
		OracleDenom:    oracleDenom,
		Redbank:        redbank,
		CreditManager:  creditManager,
		CreditAccount:  creditAccount,
		Connection:     connection,
		IBCRegistry:    ibcRegistry,
		clientRegistry: clientRegistry,
		signerAccount:  signerAccount,
		senderAddress:  senderAddress,
	}
}

// GetName returns the market identifier
func (m *MarsYieldMarket) GetName() string {
	return m.Name
}

// GetChainID returns the chain ID for the market
func (m *MarsYieldMarket) GetChainID() string {
	return m.ChainID
}

// GetDenom returns the denom of the  price for the market
func (m *MarsYieldMarket) GetDenom() string {
	return m.OracleDenom
}

// refreshMarketData ensures we have up-to-date market data
func (m *MarsYieldMarket) refreshMarketData(ctx context.Context) error {
	// If data is less than 60 seconds old, don't refresh
	if m.cachedMarket != nil && time.Since(m.lastUpdated) < 60*time.Second {
		return nil
	}

	// Fetch updated market data
	redbankClient := rb.NewQueryClient(m.Connection, m.Redbank)
	market, err := redbankClient.MarketV2(
		ctx,
		&rb.MarketV2Request{
			Denom: m.OracleDenom,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to fetch market data: %w", err)
	}

	m.cachedMarket = market
	m.lastUpdated = time.Now()
	return nil
}

// GetCurrentRate returns the current interest/liquidity rate
func (m *MarsYieldMarket) GetCurrentRate(ctx context.Context) (sdkmath.LegacyDec, error) {
	if err := m.refreshMarketData(ctx); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	rate, err := sdkmath.LegacyNewDecFromStr(m.cachedMarket.LiquidityRate)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("invalid liquidity rate: %w", err)
	}

	return rate, nil
}

// GetTotalLiquidity returns the total amount of underlying assets
func (m *MarsYieldMarket) GetTotalLiquidity(ctx context.Context) (sdkmath.Int, error) {
	if err := m.refreshMarketData(ctx); err != nil {
		return sdkmath.Int{}, err
	}

	liquidity, ok := sdkmath.NewIntFromString(m.cachedMarket.CollateralTotalAmount)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid collateral total amount")
	}

	return liquidity, nil
}

// GetTotalDebt returns the total amount of borrowed assets
func (m *MarsYieldMarket) GetTotalDebt(ctx context.Context) (sdkmath.Int, error) {
	if err := m.refreshMarketData(ctx); err != nil {
		return sdkmath.Int{}, err
	}

	debt, ok := sdkmath.NewIntFromString(m.cachedMarket.DebtTotalAmount)
	if !ok {
		return sdkmath.Int{}, fmt.Errorf("invalid debt total amount")
	}

	return debt, nil
}

// GetLentPosition returns the total amount lent including interest
func (m *MarsYieldMarket) GetLentPosition(ctx context.Context) (sdkmath.Int, error) {
	// Fetch credit positions
	if m.Connection == nil {
		return sdkmath.ZeroInt(), fmt.Errorf("connection not initialized")
	}

	// Fetch credit positions
	creditClient := cm.NewQueryClient(m.Connection, m.CreditManager)

	creditPosition, err := creditClient.Positions(
		ctx,
		&cm.PositionsRequest{
			AccountID: fmt.Sprintf("%v", m.CreditAccount),
		},
	)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to fetch credit accounts: %w", err)
	}

	// Find our lend position in this market
	for _, lend := range creditPosition.Lends {
		if lend.Denom == m.OracleDenom {
			return lend.Amount, nil
		}
	}

	return sdkmath.ZeroInt(), nil
}

// CalculateRateWithUtilization simulates the rate after changing utilization
func (m *MarsYieldMarket) CalculateRateWithUtilization(ctx context.Context, utilizationRate sdkmath.LegacyDec) (sdkmath.LegacyDec, error) {
	if err := m.refreshMarketData(ctx); err != nil {
		return sdkmath.LegacyDec{}, err
	}

	// Create a copy of the market to simulate changes
	market := m.cachedMarket

	// Get the interest rate model
	irm, err := market.InterestRateModel.ToRational()
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to convert interest rate model: %w", err)
	}

	// Calculate the new borrow rate based on utilization
	borrowRate, err := irm.GetBorrowRate(utilizationRate)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to calculate borrow rate: %w", err)
	}

	// Calculate the new liquidity rate
	reserveFactor, err := sdkmath.LegacyNewDecFromStr(market.ReserveFactor)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("invalid reserve factor: %w", err)
	}

	liquidityRate, err := irm.GetLiquidityRate(borrowRate, utilizationRate, reserveFactor)
	if err != nil {
		return sdkmath.LegacyDec{}, fmt.Errorf("failed to calculate liquidity rate: %w", err)
	}

	return liquidityRate, nil
}

// CalculateNewUtilization calculates the new utilization after adding/removing liquidity
func (m *MarsYieldMarket) CalculateNewUtilization(ctx context.Context, liquidityChange sdkmath.Int, isDeposit bool) (sdkmath.LegacyDec, error) {
	debt, err := m.GetTotalDebt(ctx)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	liquidity, err := m.GetTotalLiquidity(ctx)
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
func (m *MarsYieldMarket) MaximumWithdrawal(ctx context.Context) (sdkmath.Int, error) {
	// Get our position in this market
	creditClient := cm.NewQueryClient(m.Connection, m.CreditManager)
	creditPosition, err := creditClient.Positions(
		ctx,
		&cm.PositionsRequest{
			AccountID: fmt.Sprintf("%d", m.CreditAccount),
		},
	)
	if err != nil {
		return sdkmath.Int{}, fmt.Errorf("failed to fetch credit positions: %w", err)
	}

	// Find our lend position in this market
	for _, lend := range creditPosition.Lends {
		if lend.Denom == m.OracleDenom {
			return lend.Amount, nil
		}
	}

	// No position found
	return sdkmath.ZeroInt(), nil
}

// LendFunds deposits funds into the credit account and lends them
func (m *MarsYieldMarket) LendFunds(_ context.Context, amount sdkmath.Int) sdk.Msg {
	// Create withdrawal message
	// Prepare sender and receiver addresses
	creditAccount := fmt.Sprintf("%d", m.CreditAccount)

	amountStr := amount.String()

	// Create the actions for the credit manager
	actions := []cm.Action{}

	actions = append(actions, cm.Action{
		Deposit: &cm.Coin{
			Denom:  m.OracleDenom,
			Amount: amountStr,
		},
	})

	actions = append(actions, cm.Action{
		Lend: &cm.ActionCoin{
			Denom: m.OracleDenom,
			Amount: cm.ActionAmount{
				Exact: &amountStr,
			},
		},
	})

	depositMsg, err := cm.BuildUpdateCreditAccountMsg(
		m.senderAddress,
		m.CreditManager,
		&creditAccount,
		actions,
		sdk.Coins{sdk.NewCoin(m.OracleDenom, amount)},
	)
	if err != nil {
		return nil
	}

	return depositMsg
}

// WithdrawFunds withdraws funds from the credit account
func (m *MarsYieldMarket) WithdrawFunds(_ context.Context, amount sdkmath.Int) sdk.Msg {
	// Create withdrawal message
	// Prepare sender and receiver addresses

	creditAccount := fmt.Sprintf("%d", m.CreditAccount)

	amountStr := amount.String()

	// Create the actions for the credit manager
	actions := []cm.Action{}

	actions = append(actions, cm.Action{
		Reclaim: &cm.ActionCoin{
			Denom: m.OracleDenom,
			Amount: cm.ActionAmount{
				Exact: &amountStr,
			},
		},
	})

	actions = append(actions, cm.Action{
		WithdrawToWallet: &cm.WithdrawData{
			Coin: cm.ActionCoin{
				Denom: m.OracleDenom,
				Amount: cm.ActionAmount{
					Exact: &amountStr,
				},
			},
			Recipient: m.senderAddress, // first withdraw to the sender address
		},
	})

	// First send margin into the credit account
	withdrawMsg, err := cm.BuildUpdateCreditAccountMsg(
		m.senderAddress,
		m.CreditManager,
		&creditAccount,
		actions,
		sdk.Coins{},
	)
	if err != nil {
		// Log error and return empty slice
		return nil
	}

	return withdrawMsg
}

// TransferFunds executes a transfer between markets
func (m *MarsYieldMarket) TransferFunds(ctx context.Context, source, destination, receiver string, amount sdkmath.Int) []sdk.Msg {
	// Create withdrawal message
	// Prepare sender and receiver addresses
	client, err := m.clientRegistry.GetClient(m.ChainID, false)
	if err != nil {
		return nil
	}

	// Get the signer account
	_, sender, err := conn.GetSignerAccountAndAddress(client.Client, m.signerAccount, m.Prefix)
	if err != nil {
		return nil
	}

	coin := sdk.NewCoin(m.OracleDenom, amount)

	withdrawMsg := m.WithdrawFunds(ctx, amount)

	// Get the IBC connection for this destination
	conn, err := m.IBCRegistry.GetConnection(source, destination)
	if err != nil {
		return nil
	}

	var chainID string
	if conn.Transfer.Forward != nil {
		chainID = conn.Transfer.Forward.ChainID
	} else {
		chainID = destination
	}

	blockHeight, err := m.clientRegistry.GetHeight(ctx, chainID)
	if err != nil {
		return nil
	}

	// Create transfer message with proper IBC parameters
	transferMsg, err := ibc.CreateTransferWithMemo(
		conn.Transfer,
		source, destination,
		coin,
		uint64(*blockHeight),
		sender, receiver,
	)
	if err != nil {
		return nil
	}

	return []sdk.Msg{withdrawMsg, transferMsg}
}
