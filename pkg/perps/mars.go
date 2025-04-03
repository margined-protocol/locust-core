package perps

import (
	"context"
	"fmt"

	"github.com/margined-protocol/locust-core/pkg/contracts/mars/creditmanager"
	marsperps "github.com/margined-protocol/locust-core/pkg/contracts/mars/perps"
	"github.com/margined-protocol/locust-core/pkg/ibc"
	"github.com/margined-protocol/locust-core/pkg/types"
	"go.uber.org/zap"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abcitypes "github.com/cometbft/cometbft/abci/types"
)

const (
	MarsMaintenanceMarginRatio = 0.105 // 9.5x leverage equivalent
)

// MarsProvider implements the Provider interface for Mars Protocol
type MarsProvider struct {
	logger          *zap.Logger
	chainID         string
	config          types.MarsConfig
	collateralDenom string
	outDecimals     int

	// Providers && Clients
	msgHandler   ibc.MessageHandler
	creditClient creditmanager.QueryClient
	perpsClient  marsperps.QueryClient

	// Other
	executor string
}

// NewMarsProvider creates a new Mars provider
func NewMarsProvider(
	logger *zap.Logger,
	chainID string,
	creditClient creditmanager.QueryClient,
	perpsClient marsperps.QueryClient,
	config types.MarsConfig,
	collateralDenom string,
	outDecimals int,
	executor string,
) *MarsProvider {
	return &MarsProvider{
		logger:          logger,
		chainID:         chainID,
		creditClient:    creditClient,
		perpsClient:     perpsClient,
		config:          config,
		collateralDenom: collateralDenom,
		outDecimals:     outDecimals,
		executor:        executor,
	}
}

// Initialize implements Provider
func (m *MarsProvider) Initialize(_ context.Context) error {
	// Mars clients should already be initialized at construction
	return nil
}

// GetOpenPosition implements Provider
func (m *MarsProvider) GetPosition(ctx context.Context) (*Position, error) {
	// Fetch credit positions
	creditPosition, err := m.creditClient.Positions(
		ctx,
		&creditmanager.PositionsRequest{
			AccountID: fmt.Sprintf("%v", m.executor),
		},
	)
	if err != nil {
		m.logger.Error("Error fetching credit accounts", zap.String("executor", m.executor), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch credit accounts: %w", err)
	}

	// Fetch perp position
	position, err := m.perpsClient.Position(
		ctx,
		&marsperps.PositionRequest{
			AccountID: fmt.Sprintf("%v", m.executor),
			Denom:     m.config.Market,
		},
	)
	if err != nil {
		m.logger.Error("Error fetching perp position", zap.String("executor", m.executor), zap.Error(err))
		return nil, fmt.Errorf("failed to fetch perp position: %w", err)
	}

	// Process and get the generic position
	genericPosition, err := GetPosition(*creditPosition, position.Position, m.collateralDenom)
	if err != nil {
		m.logger.Error("Error getting position", zap.String("executor", m.executor), zap.Error(err))
		return nil, fmt.Errorf("failed to get position: %w", err)
	}

	// Convert Mars position data to the common Position struct
	return &genericPosition, nil
}

// CheckSubaccount implements Provider
func (m *MarsProvider) CheckSubaccount(account string) (bool, error) {
	// Check if credit account exists
	_, err := m.creditClient.AccountKind(context.Background(), &creditmanager.AccountKindRequest{AccountID: account})
	if err != nil {
		return false, fmt.Errorf("error fetching credit accounts: %w", err)
	}

	// As long as it doesn't error, the account exists
	return true, nil
}

// GetSubaccount implements Provider
func (m *MarsProvider) GetSubaccount() string {
	return fmt.Sprintf("%v", m.config.CreditAccount)
}

// GetProviderExecutor implements Provider
func (m *MarsProvider) GetProviderExecutor() string {
	return m.executor
}

// GetProviderDenom implements Provider
func (m *MarsProvider) GetProviderChainID() string {
	return m.chainID
}

// GetProviderName implements Provider
func (m *MarsProvider) GetProviderName() string {
	return string(ProviderMars)
}

// GetProviderDenom implements Provider
func (m *MarsProvider) GetProviderDenom() string {
	return m.config.OracleDenom
}

// GetAccountBalance implements Provider
func (m *MarsProvider) GetAccountBalance() (sdk.Coins, error) {
	return nil, fmt.Errorf("not implemented")
}

// GetSubaccountBalance implements Provider
func (m *MarsProvider) GetSubaccountBalance() (sdk.Coins, error) {
	return nil, fmt.Errorf("not implemented")
}

// CreateMarketOrder implements Provider
func (m *MarsProvider) CreateMarketOrder(_ context.Context, _, margin, size sdkmath.Int, _, reduceOnly bool) ([]sdk.Msg, error) {
	// NOTE: currently isBuy is not used but that _should_ change negative size is a sell
	// _, account, err := m.clientRegistry.GetSignerAccountAndAddress(m.signerAccount, DydxChainID)
	// if err != nil {
	// 	return nil, err
	// }
	// Price is basically unused in mars

	// nolint
	account := "todoasabove"

	// Convert the previous increasePerpPosition/decreasePerpPosition logic to handle both cases
	if reduceOnly {
		// Logic for reducing position
		return m.buildReducePositionMsgs(account, margin, size)
	}
	// Logic for increasing position
	return m.buildIncreasePositionMsgs(account, margin, size)
}

// CreateLimitOrder implements Provider
// nolint
func (m *MarsProvider) CreateLimitOrder(_ context.Context, price, margin, size sdkmath.Int, isBuy, reduceOnly bool) ([]sdk.Msg, error) {
	// Mars might not support limit orders directly
	return nil, fmt.Errorf("limit orders not supported by Mars provider")
}

// DepositSubaccount implements Provider
func (m *MarsProvider) DepositSubaccount(_ context.Context, amount sdkmath.Int) ([]sdk.Msg, error) {
	account := "todoasabove"

	m.logger.Debug("Depositing Subaccount",
		zap.String("sender", account),
		zap.String("amount", amount.String()),
	)

	creditAccount := fmt.Sprintf("%v", account)

	actions := []creditmanager.Action{}

	if amount.GT(sdkmath.ZeroInt()) {
		actions = append(actions, creditmanager.Action{
			Deposit: &creditmanager.Coin{
				Denom:  m.collateralDenom,
				Amount: amount.String(),
			},
		})
	}

	updateMsg, err := creditmanager.BuildUpdateCreditAccountMsg(
		account,
		m.config.CreditManager,
		&creditAccount,
		actions,
		sdk.NewCoins(sdk.NewCoin(m.collateralDenom, amount)),
	)
	if err != nil {
		m.logger.Error("Error creating update credit account msg", zap.Error(err))
		return nil, err
	}

	return []sdk.Msg{updateMsg}, nil
}

// WithdrawSubaccount implements Provider
func (m *MarsProvider) WithdrawSubaccount(_ context.Context, amount sdkmath.Int) ([]sdk.Msg, error) {
	account := "todoasabove"

	m.logger.Debug("Withdrawing Subaccount",
		zap.String("sender", account),
		zap.String("amount", amount.String()),
	)

	creditAccount := fmt.Sprintf("%v", account)
	amountStr := amount.Abs().String()

	actions := []creditmanager.Action{}

	if !amount.IsZero() {
		actions = append(actions, creditmanager.Action{
			WithdrawToWallet: &creditmanager.WithdrawData{
				Coin: creditmanager.ActionCoin{
					Denom: m.collateralDenom,
					Amount: creditmanager.ActionAmount{
						Exact: &amountStr,
					},
				},
				Recipient: account,
			},
		})
	}

	updateMsg, err := creditmanager.BuildUpdateCreditAccountMsg(
		account,
		m.config.CreditManager,
		&creditAccount,
		actions,
		sdk.Coins{},
	)
	if err != nil {
		m.logger.Error("Error creating update credit account msg", zap.Error(err))
		return nil, err
	}

	return []sdk.Msg{updateMsg}, nil
}

// GetLiquidationPrice implements Provider
// nolint
func (m *MarsProvider) GetLiquidationPrice(_, size, entryPrice, maintenanceMargin sdkmath.LegacyDec) sdkmath.LegacyDec {
	// Your existing liquidation price calculation
	return sdkmath.LegacyZeroDec() // Replace with actual calculation
}

// ProcessPerpEvent implements Provider
func (m *MarsProvider) ProcessPerpEvent(events []abcitypes.Event) (currentPrice string, entryPrice string, err error) {
	return ProcessMarsPerpEvent(events)
}

// CreateSubaccount implements Provider
func (m *MarsProvider) CreateSubaccount(account string) (sdk.Msg, error) {
	// Convert CreateCreditAccount to CreateSubaccount
	return creditmanager.BuildCreateCreditAccountMsg(account, m.config.CreditManager, creditmanager.AccountKind{Type: "default"})
}

// Helper functions
func (m *MarsProvider) buildIncreasePositionMsgs(account string, margin, size sdkmath.Int) ([]sdk.Msg, error) {
	m.logger.Debug("Increasing Perp Position",
		zap.String("sender", account),
		zap.String("additional_margin", margin.String()),
		zap.String("additional_amount", size.String()),
	)

	creditAccount := fmt.Sprintf("%v", account)
	orderSize := size.BigInt().String()

	actions := []creditmanager.Action{}

	if margin.GT(sdkmath.ZeroInt()) {
		actions = append(actions, creditmanager.Action{
			Deposit: &creditmanager.Coin{
				Denom:  m.collateralDenom,
				Amount: margin.String(),
			},
		})
	}

	if size.GT(sdkmath.ZeroInt()) {
		actions = append(actions, creditmanager.Action{
			ExecutePerpOrder: &creditmanager.PerpOrder{
				Denom:     m.collateralDenom,
				OrderSize: &orderSize,
			},
		})
	}

	// If there are no actions, return nil
	if len(actions) == 0 {
		return nil, nil
	}

	// First send margin into the credit account
	updateMsg, err := creditmanager.BuildUpdateCreditAccountMsg(
		account,
		m.config.CreditManager,
		&creditAccount,
		actions,
		sdk.NewCoins(sdk.NewCoin(m.collateralDenom, margin)),
	)
	if err != nil {
		m.logger.Error("Error creating update credit account msg", zap.Error(err))
		return nil, err
	}

	return []sdk.Msg{updateMsg}, nil
}

func (m *MarsProvider) buildReducePositionMsgs(account string, margin, size sdkmath.Int) ([]sdk.Msg, error) {
	m.logger.Debug("Reducing Perp Position",
		zap.String("sender", account),
		zap.String("margin_delta", margin.String()),
		zap.String("size_delta", size.String()),
	)

	creditAccount := fmt.Sprintf("%v", account)
	reduceOnly := true
	marginStr := margin.Abs().String()
	sizeStr := size.BigInt().String()

	actions := []creditmanager.Action{}

	if !size.IsZero() {
		actions = append(actions, creditmanager.Action{
			ExecutePerpOrder: &creditmanager.PerpOrder{
				Denom:      m.collateralDenom,
				OrderSize:  &sizeStr,
				ReduceOnly: &reduceOnly,
			},
		})
	}

	if !margin.IsZero() {
		actions = append(actions, creditmanager.Action{
			WithdrawToWallet: &creditmanager.WithdrawData{
				Coin: creditmanager.ActionCoin{
					Denom: m.collateralDenom,
					Amount: creditmanager.ActionAmount{
						Exact: &marginStr,
					},
				},
				Recipient: account,
			},
		})
	}

	// If there are no actions, return nil
	if len(actions) == 0 {
		return nil, nil
	}

	// First send margin into the credit account
	updateMsg, err := creditmanager.BuildUpdateCreditAccountMsg(
		account,
		m.config.CreditManager,
		&creditAccount,
		actions,
		sdk.Coins{},
	)
	if err != nil {
		m.logger.Error("Error creating update credit account msg", zap.Error(err))
		return nil, err
	}

	return []sdk.Msg{updateMsg}, nil
}

// GetPosition extracts and returns a PerpPosition from a PositionsResponse based on a given denom.
func GetPosition(creditPositions creditmanager.PositionsResponse, perpPosition *marsperps.PerpPosition, denom string) (Position, error) {
	position := Position{
		EntryPrice:    sdkmath.ZeroInt(),
		Margin:        sdkmath.ZeroInt(),
		Amount:        sdkmath.ZeroInt(),
		CurrentPrice:  sdkmath.ZeroInt(),
		UnrealizedPnl: sdkmath.ZeroInt(),
		RealizedPnl:   sdkmath.ZeroInt(),
	}

	for _, creditPosition := range creditPositions.Deposits {
		if creditPosition.Denom == denom {
			position.Margin = creditPosition.Amount
		}
	}

	if perpPosition != nil {
		amount, _ := sdkmath.NewIntFromString(*perpPosition.Size)

		position.EntryPrice = sdkmath.Int(sdkmath.LegacyMustNewDecFromStr(perpPosition.EntryPrice))
		position.CurrentPrice = sdkmath.Int(sdkmath.LegacyMustNewDecFromStr(perpPosition.CurrentPrice))
		position.Amount = amount
		position.UnrealizedPnl = sdkmath.Int(sdkmath.LegacyMustNewDecFromStr(*perpPosition.UnrealizedPnl.Pnl))
	}

	return position, nil
}

func (m *MarsProvider) IncreasePosition(ctx context.Context, _ float64, amount, _ sdkmath.Int, _ bool) (*ExecutionResult, error) {
	// Pseudo Logic
	// 1. Create message containing:
	// - Deposit to subaccount
	// - Place order
	// 2. Send message
	// 3. Return execution result

	// Get current position if any
	position, err := m.GetPosition(ctx)
	if err != nil {
		return nil, err
	}

	// Determine if we're opening a new position or increasing an existing one
	var msgs []sdk.Msg
	if position == nil || position.Amount.IsZero() {
		// New position
		msgs, err = m.CreateMarketOrder(ctx, sdkmath.ZeroInt(), amount, amount, true, false)
	} else {
		// Increase existing position
		msgs, err = m.CreateMarketOrder(ctx, sdkmath.ZeroInt(), amount, amount, position.Amount.IsPositive(), false)
	}

	if err != nil {
		return nil, err
	}

	result := &ExecutionResult{
		Messages: msgs,
		Position: position,
	}

	// If handler is provided, execute the transaction
	if m.msgHandler != nil {
		resp, err := m.msgHandler(m.chainID, msgs, false, true)
		if err != nil {
			return result, err
		}

		result.TxHash = resp.TxHash
		result.Events = resp.Events
		result.Executed = true

		// Get updated position after execution
		updatedPosition, err := m.GetPosition(ctx)
		if err != nil {
			return result, err
		}
		result.Position = updatedPosition
	}

	return result, nil
}

// nolint
func (m *MarsProvider) ReducePosition(_ context.Context, price float64, amount, margin sdkmath.Int, isLong bool) (*ExecutionResult, error) {
	return nil, fmt.Errorf("not implemented")
}

// nolint
func (m *MarsProvider) ClosePosition(ctx context.Context, isLong bool) (*ExecutionResult, error) {
	return nil, fmt.Errorf("not implemented")
}

// nolint
func (m *MarsProvider) AdjustMargin(ctx context.Context, margin sdkmath.Int, isAdd bool) (*ExecutionResult, error) {
	return nil, fmt.Errorf("not implemented")
}
