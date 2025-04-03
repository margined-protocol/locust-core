package market

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// These are used in levana messages
const (
	Long  = "long"
	Short = "short"
)

// ExecuteOwnerMsg represents owner-related execution messages
type ExecuteOwnerMsg struct {
	ConfigUpdate *ConfigUpdateMsg `json:"config_update,omitempty"`
}

// ConfigUpdateMsg updates market configuration
type ConfigUpdateMsg struct {
	Update ConfigUpdate `json:"update"`
}

// SetManualPriceMsg sets a manual spot price
type SetManualPriceMsg struct {
	Price    float64 `json:"price,string"`
	PriceUSD float64 `json:"price_usd,string"`
}

// ReceiveMsg represents a CW20 receive message
type ReceiveMsg struct {
	Amount string `json:"amount"`
	Msg    string `json:"msg"`
	Sender string `json:"sender"`
}

// OpenPositionMsg represents opening a new position
type OpenPositionMsg struct {
	SlippageAssert   *SlippageAssert `json:"slippage_assert,omitempty"`
	Leverage         float64         `json:"leverage,string"`
	Direction        string          `json:"direction"`
	MaxGains         *float64        `json:"max_gains,string,omitempty"`
	StopLossOverride *float64        `json:"stop_loss_override,string,omitempty"`
	TakeProfit       *float64        `json:"take_profit,string,omitempty"`
	Amount           int64           `json:"amount,string"`
}

// UpdatePositionAddCollateralImpactLeverageMsg adds collateral while impacting leverage
type UpdatePositionAddCollateralImpactLeverageMsg struct {
	ID int `json:"id,string"`
}

// UpdatePositionRemoveCollateralImpactLeverageMsg removes collateral while impacting leverage
type UpdatePositionRemoveCollateralImpactLeverageMsg struct {
	ID     int     `json:"id,string"`
	Amount float64 `json:"amount,string"`
}

// UpdatePositionRemoveCollateralImpactSizeMsg represents the message structure for removing collateral
type UpdatePositionRemoveCollateralImpactSizeMsg struct {
	ID             int             `json:"id,string"`
	Amount         float64         `json:"amount,string"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

// UpdatePositionRemoveCollateralImpactLeverageMsg represents the structure for removing collateral and increasing leverage
type UpdatePositionAddCollateralImpactSizeMsg struct {
	ID             int             `json:"id,string"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

// UpdatePositionLeverageMsg represents changing leverage
type UpdatePositionLeverageMsg struct {
	ID             int             `json:"id,string"`
	Leverage       float64         `json:"leverage,string"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

// UpdatePositionMaxGainsMsg updates max gains setting
type UpdatePositionMaxGainsMsg struct {
	ID       int      `json:"id,string"`
	MaxGains *float64 `json:"max_gains,string,omitempty"`
}

// UpdatePositionTakeProfitPriceMsg represents the message structure for updating the take profit price
type UpdatePositionTakeProfitPriceMsg struct {
	ID    int     `json:"id,string"`
	Price float64 `json:"price,string"`
}

// StopLoss represents the stop-loss price inside UpdatePositionStopLossPriceMsg
type StopLoss struct {
	Price float64 `json:"price,string"`
}

// UpdatePositionStopLossPriceMsg sets the stop-loss price
type UpdatePositionStopLossPriceMsg struct {
	ID       int      `json:"id,string"`
	StopLoss StopLoss `json:"stop_loss"`
}

// SetTriggerOrderMsg sets a trigger order
type SetTriggerOrderMsg struct {
	ID               int     `json:"id,string"`
	StopLossOverride *string `json:"stop_loss_override,omitempty"`
	TakeProfit       *string `json:"take_profit,omitempty"`
}

// PlaceLimitOrderMsg places a new limit order
type PlaceLimitOrderMsg struct {
	TriggerPrice     float64  `json:"trigger_price,string"`
	Leverage         float64  `json:"leverage,string"`
	Direction        string   `json:"direction"`
	MaxGains         *float64 `json:"max_gains,string,omitempty"`
	StopLossOverride *float64 `json:"stop_loss_override,string,omitempty"`
	TakeProfit       *float64 `json:"take_profit,string,omitempty"`
}

// CancelLimitOrderMsg cancels a limit order
type CancelLimitOrderMsg struct {
	OrderID int `json:"order_id,string"`
}

// ClosePositionMsg represents closing a position
type ClosePositionMsg struct {
	ID             int             `json:"id,string"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

// CrankMsg represents executing a crank function
type CrankMsg struct {
	Execs   *int    `json:"execs,omitempty"`
	Rewards *string `json:"rewards,omitempty"`
}

// DepositLiquidityMsg handles liquidity deposits
type DepositLiquidityMsg struct {
	StakeToXLP bool `json:"stake_to_xlp"`
}

// ReinvestYieldMsg reinvests yield into XLP
type ReinvestYieldMsg struct {
	StakeToXLP bool `json:"stake_to_xlp"`
}

// WithdrawLiquidityMsg withdraws liquidity
type WithdrawLiquidityMsg struct {
	LPAmount float64 `json:"lp_amount"`
}

// ClaimYieldMsg allows users to claim yield
type ClaimYieldMsg struct{}

// StakeLPMsg allows staking LP tokens
type StakeLPMsg struct {
	// Amount string `json:"amount"`
}

// UnstakeXLPMsg unstake XLP tokens
type UnstakeXLPMsg struct {
	// Amount string `json:"amount"`
}

// StopUnstakingXLPMsg stops an ongoing unstaking of XLP
type StopUnstakingXLPMsg struct{}

// CollectUnstakedLPMsg collects unstaked LP tokens
type CollectUnstakedLPMsg struct{}

// NFTProxyMsg proxies an NFT execution
type NFTProxyMsg struct {
	Sender string `json:"sender"`
	Msg    string `json:"msg"`
}

// LiquidityTokenProxyMsg proxies a liquidity token execution
type LiquidityTokenProxyMsg struct {
	Sender string `json:"sender"`
	Kind   string `json:"kind"`
	Msg    string `json:"msg"`
}

// TransferDaoFeesMsg transfers DAO fees
type TransferDaoFeesMsg struct{}

// CloseAllPositionsMsg closes all open positions
type CloseAllPositionsMsg struct{}

// ProvideCrankFundsMsg funds the crank system
type ProvideCrankFundsMsg struct {
	// Amount types.Coins `json:"amount"`
}

// PerformDeferredExecMsg performs a deferred execution
type PerformDeferredExecMsg struct {
	ID                  int    `json:"id,string"`
	PricePointTimestamp string `json:"price_point_timestamp"`
}

// ConfigUpdate represents config updates
type ConfigUpdate struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SlippageAssert represents the optional slippage assertion for the update position message
type SlippageAssert struct {
	Price     float64 `json:"price,string"`
	Tolerance float64 `json:"tolerance,string"`
}

// CreateCrankMsg creates a crank message
// Deprecated: to remove (used in levanacrank)
func CreateCrankMsg(contractAddress string, senderAddress string) (*wasmtypes.MsgExecuteContract, error) {
	jsonData, err := json.Marshal(map[string]any{"crank": struct{}{}})
	if err != nil {
		return nil, err
	}

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   senderAddress,
		Contract: contractAddress,
		Msg:      jsonData,
		Funds:    nil,
	}

	return msg, nil
}
