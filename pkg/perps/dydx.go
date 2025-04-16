package perps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"github.com/margined-protocol/locust/core/pkg/connection"
	"github.com/margined-protocol/locust/core/pkg/grpc"
	"github.com/margined-protocol/locust/core/pkg/ibc"
	"github.com/margined-protocol/locust/core/pkg/math"
	clob "github.com/margined-protocol/locust/core/pkg/proto/dydx/clob/types"
	send "github.com/margined-protocol/locust/core/pkg/proto/dydx/sending/types"
	subaccounts "github.com/margined-protocol/locust/core/pkg/proto/dydx/subaccounts/types"

	"go.uber.org/zap"
)

const (
	DydxChainID        = "dydx-mainnet-1"
	DefaultOrderExpiry = 60   // seconds
	DefaultSlippage    = 0.01 // 1% - eventually we should use the best bid/ask or make market orders idk
	maxRetries         = 3
	retryDelay         = 2 * time.Second
)

// DydxProvider implements the Provider interface for Dydx Protocol
type DydxProvider struct {
	logger                    *zap.Logger
	BaseChainID               string
	assetId                   uint32 // USDC == 0
	marketId                  uint32
	subaccountId              uint32
	market                    string
	denom                     string
	atomicResolution          int64  // Decimal places for the amount, -6 == 6dp
	quantumConversionExponent int64  // Decimal places for the price, -9 == 9dp
	decimals                  int64  // Decimal places for the asset used externally
	stepBaseQuantums          uint64 // MinQuantityTickSize
	subticksPerTick           uint64 // MinPriceTickSize
	minEquity                 sdkmath.Int

	// Providers && Clients
	clientRegistry   *connection.ClientRegistry
	subaccountClient subaccounts.QueryClient
	msgHandler       ibc.MessageHandler

	// Dydx Indexer
	indexerURL string
	httpClient *http.Client

	signerAccount string
	executor      string

	// Defaults (TODO: make these configurable)
	slippage    float64
	orderExpiry uint64 // Order expiry in seconds
}

type DydxConfig struct {
	Market           string
	MarketID         uint32
	SubticksPerTick  uint64
	StepBaseQuantums uint64
	MinEquity        sdkmath.Int

	SignerAccount string
	Executor      string
	SubaccountID  uint32
	Denom         string

	BaseChainID string

	SubaccountClient subaccounts.QueryClient
	ClientRegistry   *connection.ClientRegistry
	MsgHandler       ibc.MessageHandler
	IndexerURL       string

	QuantumConversionExp int64
	AtomicResolution     int64
	Decimals             int64
}

// NewDydxProvider creates a new Dydx provider
func NewDydxProvider(
	// Logger
	logger *zap.Logger,

	// Market configuration
	marketId uint32,
	market string,
	subticksPerTick uint64,
	stepBaseQuantums uint64,
	quantumConversionExponent int64,
	atomicResolution int64,
	decimals int64,
	minEquity sdkmath.Int,

	// Account configuration
	signerAccount string,
	executor string,
	subaccountId uint32,
	denom string,

	// Chain configuration
	baseChainID string,

	// Clients and connections
	subaccountClient subaccounts.QueryClient,
	clientRegistry *connection.ClientRegistry,
	msgHandler ibc.MessageHandler,
	indexerURL string,
) *DydxProvider {
	// Initialize HTTP client with reasonable defaults
	httpClient := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	return &DydxProvider{
		// Logger
		logger: logger,

		// Market configuration
		marketId:                  marketId,
		market:                    market,
		subticksPerTick:           subticksPerTick,
		stepBaseQuantums:          stepBaseQuantums,
		quantumConversionExponent: quantumConversionExponent,
		atomicResolution:          atomicResolution,
		slippage:                  DefaultSlippage,
		minEquity:                 minEquity,
		decimals:                  decimals,

		// Account configuration
		signerAccount: signerAccount,
		executor:      executor,
		subaccountId:  subaccountId,
		denom:         denom,

		// Chain configuration
		BaseChainID: baseChainID,
		orderExpiry: DefaultOrderExpiry,

		// Clients and connections
		subaccountClient: subaccountClient,
		clientRegistry:   clientRegistry,
		msgHandler:       msgHandler,
		indexerURL:       indexerURL,
		httpClient:       httpClient,
	}
}

// Initialize implements Provider
func (m *DydxProvider) Initialize(ctx context.Context) error {
	// Dydx clients should already be initialized at construction
	return nil
}

// GetPosition implements Provider
func (m *DydxProvider) GetPosition(ctx context.Context) (*Position, error) {
	m.logger.Info("Getting position", zap.Any("market", m.market))
	_, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	// Query subaccount information from indexer first
	result, err := m.QuerySubaccountIndexer(ctx, account, m.subaccountId)
	if err != nil {
		return nil, fmt.Errorf("error fetching indexer data: %w", err)
	}

	candles, err := m.QueryCandlePrices(ctx, m.market)
	if err != nil {
		return nil, fmt.Errorf("error fetching candle prices: %w", err)
	}

	position, err := ProcessIndexerResponse(m.market, m.decimals, result)
	if err != nil {
		return nil, fmt.Errorf("error processing indexer data: %w", err)
	}

	currentPrice, err := ProcessCandlesResponse(m.market, candles)
	if err != nil {
		return nil, fmt.Errorf("error processing candle prices: %w", err)
	}

	position.CurrentPrice = *currentPrice

	return position, nil

}

// CheckSubaccount implements Provider
func (m *DydxProvider) CheckSubaccount(account string) (bool, error) {
	// No need to check subaccount they exist by default
	return true, nil
}

// GetSubaccount implements Provider
func (m *DydxProvider) GetSubaccount() string {
	return fmt.Sprintf("%d", m.subaccountId)
}

// GetProviderDenom implements Provider
func (m *DydxProvider) GetProviderChainID() string {
	return DydxChainID
}

// GetProviderName implements Provider
func (m *DydxProvider) GetProviderName() string {
	return string(ProviderDydx)
}

// GetProviderDenom implements Provider
func (m *DydxProvider) GetProviderDenom() string {
	return m.denom
}

// GetProviderExecutor implements Provider
func (m *DydxProvider) GetProviderExecutor() string {
	return m.executor
}

// GetAccountBalance implements Provider
func (m *DydxProvider) GetAccountBalance() (sdk.Coins, error) {
	cl, err := m.clientRegistry.GetClient(DydxChainID, false)
	if err != nil {
		return nil, err
	}

	dydxConn, err := grpc.SetupGRPCConnection(cl.Chain.GRPCServerAddress, cl.Chain.GRPCTLS)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GRPC server: %w", err)
	}

	bankClient := banktypes.NewQueryClient(dydxConn)

	_, signer, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	res, err := bankClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: signer,
	})
	if err != nil {
		return nil, err
	}

	return res.Balances, nil
}

// GetSubaccountBalance implements Provider
func (m *DydxProvider) GetSubaccountBalance() (sdk.Coins, error) {
	_, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	// Query subaccount information from indexer first
	result, err := m.QuerySubaccountIndexer(context.Background(), account, m.subaccountId)
	if err != nil {
		return nil, fmt.Errorf("error fetching indexer data: %w", err)
	}

	// We use equity as a proxy for the subaccount balance
	equity := result.Subaccount.Equity

	// Parse the equity string into a decimal
	equityDec, err := sdkmath.LegacyNewDecFromStr(equity)
	if err != nil {
		return nil, fmt.Errorf("failed to parse equity value %s: %w", equity, err)
	}

	// Convert to integer with 6 decimals precision (assuming USDC)
	equityInt := equityDec.MulInt64(1_000_000).TruncateInt()

	coins := sdk.NewCoins(sdk.NewCoin(m.denom, equityInt))

	return coins, nil
}

// CreateMarketOrder implements Provider
func (m *DydxProvider) CreateMarketOrder(ctx context.Context, price, margin, size sdkmath.Int, isBuy, reduceOnly bool) ([]sdk.Msg, error) {

	_, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	height, err := m.clientRegistry.GetHeight(ctx, DydxChainID)
	if err != nil {
		return nil, err
	}

	// Validate non-zero inputs
	// NOTE: we do not error here
	if price.IsZero() {
		return nil, nil
	}
	if size.IsZero() {
		return nil, nil
	}

	m.logger.Info("Creating market order",
		zap.String("price", price.String()),
		zap.String("size", size.String()),
		zap.Bool("isBuy", isBuy),
		zap.Bool("reduceOnly", reduceOnly),
	)

	// Validate and round the price
	validPrice, err := m.validateAndRoundPrice(price)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	// Validate and round the size
	validSize, err := m.validateAndRoundAmount(size)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	side := clob.Order_Side_value["SIDE_BUY"]
	if !isBuy {
		side = clob.Order_Side_value["SIDE_SELL"]
	}

	quantums := validSize.Uint64()
	subticks := validPrice.Uint64()

	order := &clob.MsgPlaceOrder{
		Order: clob.Order{
			OrderId: clob.OrderId{
				SubaccountId: subaccounts.SubaccountId{
					Owner: account,
				},
				ClobPairId: m.marketId,
				OrderFlags: 0, // 0 short-term, 32 conditional, 64 long-term
			},
			Side:       clob.Order_Side(side),
			Quantums:   quantums,
			Subticks:   subticks,
			ReduceOnly: reduceOnly,
			GoodTilOneof: &clob.Order_GoodTilBlock{
				GoodTilBlock: uint32(*height + 10),
			},
			TimeInForce: clob.Order_TimeInForce(clob.Order_TimeInForce_value["TIME_IN_FORCE_IOC"]),
		},
	}

	m.logger.Debug("Market order created successfully",
		zap.Uint64("quantums", quantums),
		zap.Uint64("subticks", subticks),
		zap.String("account", account),
	)

	return []sdk.Msg{order}, nil
}

// CreateLimitOrder implements Provider
func (m *DydxProvider) CreateLimitOrder(ctx context.Context, price, margin, size sdkmath.Int, isBuy, reduceOnly bool) ([]sdk.Msg, error) {
	_, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	// Validate non-zero inputs
	// NOTE: we do not error here
	if price.IsZero() {
		return nil, nil
	}
	if size.IsZero() {
		return nil, nil
	}

	m.logger.Info("Creating market order",
		zap.String("price", price.String()),
		zap.String("size", size.String()),
		zap.Bool("isBuy", isBuy),
		zap.Bool("reduceOnly", reduceOnly),
	)

	// Validate and round the price
	validPrice, err := m.validateAndRoundPrice(price)
	if err != nil {
		return nil, fmt.Errorf("invalid price: %w", err)
	}

	// Validate and round the size
	validSize, err := m.validateAndRoundAmount(size)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	side := clob.Order_Side_value["SIDE_BUY"]
	if !isBuy {
		side = clob.Order_Side_value["SIDE_SELL"]
	}

	quantums := validSize.Uint64()
	subticks := validPrice.Uint64()

	expiryBlockTime := time.Now().Add(time.Duration(m.orderExpiry) * time.Second)

	order := &clob.MsgPlaceOrder{
		Order: clob.Order{
			OrderId: clob.OrderId{
				SubaccountId: subaccounts.SubaccountId{
					Owner: account,
				},
				ClobPairId: m.marketId,
				OrderFlags: 64, // 0 short-term, 32 conditional, 64 long-term
			},
			Side:       clob.Order_Side(side),
			Quantums:   quantums,
			Subticks:   subticks,
			ReduceOnly: reduceOnly,
			GoodTilOneof: &clob.Order_GoodTilBlockTime{
				GoodTilBlockTime: uint32(expiryBlockTime.Unix()),
			},
		},
	}

	m.logger.Debug("Limit order created successfully",
		zap.Uint64("quantums", quantums),
		zap.Uint64("subticks", subticks),
		zap.String("account", account),
	)

	return []sdk.Msg{order}, nil
}

// DepositSubaccount implements Provider
func (m *DydxProvider) DepositSubaccount(ctx context.Context, amount sdkmath.Int) ([]sdk.Msg, error) {
	// Validate non-zero amount
	if amount.IsZero() {
		m.logger.Info("Skipping deposit for zero amount")
		return nil, nil
	}

	_, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	deposit := send.MsgDepositToSubaccount{
		Sender: account,
		Recipient: subaccounts.SubaccountId{
			Owner:  account,
			Number: m.subaccountId,
		},
		AssetId:  m.assetId,
		Quantums: amount.Uint64(),
	}
	return []sdk.Msg{&deposit}, nil
}

// WithdrawSubaccount implements Provider
func (m *DydxProvider) WithdrawSubaccount(ctx context.Context, amount sdkmath.Int) ([]sdk.Msg, error) {
	// Validate non-zero amount early
	if amount.IsZero() {
		m.logger.Info("Skipping withdrawal for zero amount")
		return nil, nil
	}

	// Get subaccount balance
	balance, err := m.GetSubaccountBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get subaccount balance: %w", err)
	}

	// Find USDC balance
	equity := sdkmath.ZeroInt()
	for _, coin := range balance {
		if coin.Denom == m.denom {
			equity = coin.Amount
			break
		}
	}

	// Equity is the sum of the margin and the USDC balance, noting that margin is USDC
	m.logger.Info("Equity", zap.Any("equity", equity))

	// Calculate maximum withdrawal amount
	maxWithdrawal := equity.Sub(m.minEquity)
	if maxWithdrawal.IsNegative() {
		m.logger.Warn("Cannot withdraw: balance below minimum equity",
			zap.String("balance", equity.String()),
			zap.String("min_equity", m.minEquity.String()),
		)
		return nil, nil
	}

	// Adjust withdrawal amount if needed
	if amount.GT(maxWithdrawal) {
		m.logger.Warn("Adjusting withdrawal amount to maintain minimum equity",
			zap.String("requested", amount.String()),
			zap.String("adjusted", maxWithdrawal.String()),
		)
		amount = maxWithdrawal
	}

	// Skip if nothing to withdraw after adjustment
	if amount.IsZero() {
		m.logger.Info("Skipping withdrawal: adjusted amount is zero")
		return nil, nil
	}

	// Get signer account address
	_, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get signer account: %w", err)
	}

	// Create withdrawal message
	withdraw := &send.MsgWithdrawFromSubaccount{
		Sender: subaccounts.SubaccountId{
			Owner:  account,
			Number: m.subaccountId,
		},
		Recipient: account,
		AssetId:   m.assetId,
		Quantums:  amount.Uint64(),
	}

	return []sdk.Msg{withdraw}, nil
}

// GetLiquidationPrice implements Provider
func (m *DydxProvider) GetLiquidationPrice(equity, size, entryPrice, maintenanceMargin sdkmath.LegacyDec) sdkmath.LegacyDec {
	// If no position, return zero
	if size.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	// Calculate liquidation price using the formula:
	// p' = (e - s * p) / (|s| * MMF - s)

	// Calculate numerator: (e - s * p)
	numerator := equity.Sub(size.Mul(entryPrice))

	// Calculate denominator: (|s| * MMF - s)
	denominator := size.Abs().Mul(maintenanceMargin).Sub(size)

	// Handle division by zero
	if denominator.IsZero() {
		return sdkmath.LegacyZeroDec()
	}

	// Calculate the liquidation price
	liquidationPrice := numerator.Quo(denominator)

	// Ensure liquidation price is not negative
	if liquidationPrice.IsNegative() {
		return sdkmath.LegacyZeroDec()
	}

	return liquidationPrice
}

// ProcessPerpEvent implements Provider
func (m *DydxProvider) ProcessPerpEvent(_ []abcitypes.Event) (currentPrice string, entryPrice string, err error) {
	// Get fills from indexer
	fills, err := m.QueryFillsIndexer(context.Background(), m.executor, m.subaccountId)
	if err != nil {
		return "", "", fmt.Errorf("error fetching fills: %w", err)
	}

	// Check if fills response is empty
	if fills == nil || len(fills.Fills) == 0 {
		m.logger.Warn("No fills found in indexer response")
		return "", "", fmt.Errorf("no fills available")
	}

	// Get entry price from most recent fill
	entryPrice = fills.Fills[0].Price
	if entryPrice == "" {
		m.logger.Warn("Fill has empty price")
		return "", "", fmt.Errorf("fill price is empty")
	}

	// Get current price from candles
	candles, err := m.QueryCandlePrices(context.Background(), m.market)
	if err != nil {
		return "", "", fmt.Errorf("error fetching candles: %w", err)
	}

	// Check if candles response is empty
	if candles == nil || len(candles.Candles) == 0 {
		m.logger.Warn("No candles found in response")
		return "", entryPrice, fmt.Errorf("no candles available")
	}

	// Get current price from most recent candle
	currentPrice = candles.Candles[0].Close
	if currentPrice == "" {
		m.logger.Warn("Candle has empty close price")
		return "", entryPrice, fmt.Errorf("candle close price is empty")
	}

	m.logger.Debug("Successfully processed perp event",
		zap.String("market", m.market),
		zap.String("currentPrice", currentPrice),
		zap.String("entryPrice", entryPrice),
	)

	return currentPrice, entryPrice, nil
}

// CreateSubaccount implements Provider
func (m *DydxProvider) CreateSubaccount(account string) (sdk.Msg, error) {
	return nil, fmt.Errorf("not implemented")
}

// Helper functions
func (m *DydxProvider) IncreasePosition(ctx context.Context, price float64, amount, margin sdkmath.Int, isLong bool) (*ExecutionResult, error) {
	// Validate margin is not negative
	if margin.IsNegative() {
		return nil, fmt.Errorf("margin cannot be negative")
	}

	// High level-logic:
	// 1. Create a deposit message
	// 2. Create a place order message
	// 3. Send the messages

	// 1. Create a deposit message
	depositMsgs, err := m.DepositSubaccount(ctx, margin)
	if err != nil {
		return nil, err
	}

	// 2. Send the messages - MsgPlaceOrder, MsgCancel, MsgBatchCancel cannot have multiple messages in a single tx
	if depositMsgs != nil {
		depositTx, err := m.msgHandler(DydxChainID, depositMsgs, false, false)
		if err != nil {
			return nil, err
		}
		m.logger.Info("Deposit transaction", zap.Any("tx", depositTx.TxHash))
	}

	adjustedPrice := math.AdjustSlippageFloat64(price, m.slippage, isLong)
	quantumPrice := math.FloatToQuantumPrice(adjustedPrice, m.quantumConversionExponent)

	m.logger.Info("Quantum price", zap.Any("price", quantumPrice))
	m.logger.Info("Subticks per tick", zap.Any("price", m.subticksPerTick))
	quantumPrice = math.RoundFixedPointInt(quantumPrice, m.subticksPerTick)

	m.logger.Info("Original price", zap.Any("price", price))
	m.logger.Info("Quantum price", zap.Any("price", quantumPrice))

	// 2. Create a place order message
	orderMsgs, err := m.CreateLimitOrder(ctx, quantumPrice, margin, amount, isLong, false)
	if err != nil {
		return nil, err
	}

	// Get initial position
	initialPosition, err := m.GetPosition(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial position: %w", err)
	}

	var orderTx *cosmosclient.Response
	// 3. Send the messages with retry mechanism
	if orderMsgs != nil {
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			var err error
			orderTx, err = m.msgHandler(DydxChainID, orderMsgs, true, false)
			if err != nil {
				return nil, err
			}

			m.logger.Info("Order transaction", zap.Any("tx", orderTx.TxHash), zap.Int("attempt", attempt))

			// Verify the position change
			err = verifyPositionChange(ctx, m.logger, m, initialPosition, amount)
			if err == nil {
				// Success - position verified
				break
			}

			if attempt < maxRetries {
				m.logger.Warn("Position verification failed, retrying...",
					zap.Int("attempt", attempt),
					zap.Int("max_retries", maxRetries),
					zap.Error(err))
				// Wait a bit before retrying
				time.Sleep(time.Second * 2)
				continue
			}

			// If we've exhausted all retries, return the error
			return nil, fmt.Errorf("failed to verify position increase after %d attempts: %w", maxRetries, err)
		}
	}

	result := &ExecutionResult{
		Messages: append(depositMsgs, orderMsgs...),
	}

	if orderTx != nil {
		result.TxHash = orderTx.TxHash
		result.Events = orderTx.Events
	}

	return result, nil
}

func (m *DydxProvider) ReducePosition(ctx context.Context, price float64, amount, margin sdkmath.Int, isLong bool) (*ExecutionResult, error) {
	// High level-logic:
	// 1. Create a reduce position message
	// 2. Send the message
	adjustedPrice := math.AdjustSlippageFloat64(price, m.slippage, isLong)
	quantumPrice := math.FloatToQuantumPrice(adjustedPrice, m.quantumConversionExponent)

	quantumPrice = math.RoundFixedPointInt(quantumPrice, m.subticksPerTick)

	// 1. Create a reduce order message only market orders can be reduce only
	orderMsgs, err := m.CreateMarketOrder(ctx, quantumPrice, margin, amount, isLong, true)
	if err != nil {
		return nil, err
	}

	// Use fee client as these messages are short-term
	client, err := m.clientRegistry.GetClient(DydxChainID, true)
	if err != nil {
		return nil, err
	}

	account, _, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("order msgs", zap.Any("msgs", orderMsgs))

	// 2. Send the messages
	if orderMsgs != nil {
		m.logger.Info("Sending short term order")
		err = m.sendShortTermOrder(ctx, client.Client, account, orderMsgs, amount.Neg())
		if err != nil {
			return nil, err
		}
	}

	// 1. Create a withdraw message
	withdrawMsgs, err := m.WithdrawSubaccount(ctx, margin)
	if err != nil {
		return nil, err
	}

	m.logger.Debug("withdraw msgs", zap.Any("msgs", withdrawMsgs))

	// 2. Send the messages - MsgPlaceOrder, MsgCancel, MsgBatchCancel cannot have multiple messages in a single tx
	var orderTx *cosmosclient.Response
	if withdrawMsgs != nil {
		orderTx, err = m.msgHandler(DydxChainID, withdrawMsgs, false, false)
		if err != nil {
			return nil, err
		}
	}

	result := &ExecutionResult{
		Messages: append(orderMsgs, withdrawMsgs...),
	}

	if orderTx != nil {
		result.TxHash = orderTx.TxHash
		result.Events = orderTx.Events
	}

	return result, nil

}

func (m *DydxProvider) ClosePosition(ctx context.Context, isLong bool) (*ExecutionResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *DydxProvider) AdjustMargin(ctx context.Context, margin sdkmath.Int, isAdd bool) (*ExecutionResult, error) {
	return nil, fmt.Errorf("not implemented")
}

// sendShortTermOrder sends a short-term order and verifies its execution by checking position changes
func (m *DydxProvider) sendShortTermOrder(
	ctx context.Context, client *cosmosclient.Client, account *cosmosaccount.Account,
	orderMsgs []sdk.Msg, expectedSizeChange sdkmath.Int,
) error {
	// Get initial position
	initialPosition, err := m.GetPosition(ctx)
	if err != nil {
		return fmt.Errorf("failed to get initial position: %w", err)
	}

	// Broadcast the order
	err = connection.BroadcastShortTermOrder(ctx, m.logger, client, *account, orderMsgs...)
	if err != nil {
		return fmt.Errorf("failed to broadcast order: %w", err)
	}

	// Wait a short time for the order to be processed
	time.Sleep(1 * time.Second)

	// Check position multiple times to confirm the change
	for attempts := 0; attempts < 3; attempts++ {
		newPosition, err := m.GetPosition(ctx)
		if err != nil {
			m.logger.Warn("Failed to get updated position",
				zap.Error(err),
				zap.Int("attempt", attempts+1),
			)
			time.Sleep(1 * time.Second)
			continue
		}

		// Calculate actual size change
		actualChange := newPosition.Amount.Sub(initialPosition.Amount)

		// Check if the position changed as expected
		if actualChange.Equal(expectedSizeChange) {
			m.logger.Info("Order executed successfully",
				zap.String("initial_size", initialPosition.Amount.String()),
				zap.String("final_size", newPosition.Amount.String()),
				zap.String("change", actualChange.String()),
			)
			return nil
		}

		m.logger.Debug("Position not yet updated",
			zap.String("expected_change", expectedSizeChange.String()),
			zap.String("actual_change", actualChange.String()),
			zap.Int("attempt", attempts+1),
		)

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("order execution not confirmed: expected size change %s not observed", expectedSizeChange)
}

// validateAndRoundPrice ensures the price is a multiple of subticksPerTick
func (m *DydxProvider) validateAndRoundPrice(price sdkmath.Int) (sdkmath.Int, error) {
	if price.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("price cannot be negative")
	}

	roundedPrice := math.RoundFixedPointInt(price, m.subticksPerTick)
	return roundedPrice, nil
}

// validateAndRoundAmount ensures the amount is >= stepBaseQuantums and is a multiple thereof
func (m *DydxProvider) validateAndRoundAmount(amount sdkmath.Int) (sdkmath.Int, error) {
	if amount.IsNegative() {
		return sdkmath.Int{}, fmt.Errorf("amount cannot be negative")
	}

	minAmount := sdkmath.NewInt(int64(m.stepBaseQuantums))
	if amount.LT(minAmount) {
		return sdkmath.Int{}, fmt.Errorf("amount %s is less than minimum allowed %s", amount, minAmount)
	}

	roundedAmount := math.RoundFixedPointInt(amount, m.stepBaseQuantums)
	return roundedAmount, nil
}

// Add these new methods
func (m *DydxProvider) QueryIndexer(ctx context.Context, path string) (map[string]interface{}, error) {
	if m.indexerURL == "" {
		return nil, fmt.Errorf("indexer URL not configured")
	}

	url := fmt.Sprintf("%s%s", m.indexerURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indexer request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// Helper method for subaccount queries
func (m *DydxProvider) QueryFillsIndexer(ctx context.Context, address string, subaccountNumber uint32) (*IndexerFillResponse, error) {
	path := fmt.Sprintf("/fills?address=%s&subaccountNumber=%d", address, subaccountNumber)

	url := fmt.Sprintf("%s%s", m.indexerURL, path)

	m.logger.Info("Querying fills indexer", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indexer request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result IndexerFillResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Helper method for subaccount queries
func (m *DydxProvider) QuerySubaccountIndexer(ctx context.Context, address string, subaccountNumber uint32) (*IndexerSubaccountResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		result, err := m.querySubaccountIndexer(ctx, address, subaccountNumber)
		if err == nil {
			return result, nil
		}
		lastErr = err
		m.logger.Warn("Retrying indexer query",
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		time.Sleep(retryDelay)
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// Move existing query logic to private method
func (m *DydxProvider) querySubaccountIndexer(ctx context.Context, address string, subaccountNumber uint32) (*IndexerSubaccountResponse, error) {
	path := fmt.Sprintf("/addresses/%s/subaccountNumber/%d", address, subaccountNumber)

	url := fmt.Sprintf("%s%s", m.indexerURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indexer request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result IndexerSubaccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// Helper method for subaccount queries
func (m *DydxProvider) QueryCandlePrices(ctx context.Context, market string) (*IndexerCandleResponse, error) {
	path := fmt.Sprintf("/candles/perpetualMarkets/%s?resolution=5MINS", market)

	url := fmt.Sprintf("%s%s", m.indexerURL, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indexer request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result IndexerCandleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
