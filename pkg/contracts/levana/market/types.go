package market

import (
	"encoding/json"

	"github.com/margined-protocol/locust-core/pkg/time"
)

type StatusRequest struct{}

// StatusResponse represents data on the market
type StatusResponse struct {
	MarketID                  string     `json:"market_id"`
	Base                      string     `json:"base"`
	Quote                     string     `json:"quote"`
	MarketType                string     `json:"market_type"`
	Collateral                Collateral `json:"collateral"`
	Config                    Config     `json:"config"`
	Liquidity                 Liquidity  `json:"liquidity"`
	NextCrank                 *string    `json:"next_crank"`
	LastCrank                 string     `json:"last_crank_completed"`
	NextDeferred              *string    `json:"next_deferred_execution"`
	NewestDeferred            *string    `json:"newest_deferred_execution"`
	NextLiquifunding          string     `json:"next_liquifunding"`
	DeferredItems             int        `json:"deferred_execution_items"`
	LastExecID                string     `json:"last_processed_deferred_exec_id"`
	BorrowFee                 string     `json:"borrow_fee"`
	BorrowFeeLP               string     `json:"borrow_fee_lp"`
	BorrowFeeXLP              string     `json:"borrow_fee_xlp"`
	LongFunding               string     `json:"long_funding"`
	ShortFunding              string     `json:"short_funding"`
	LongNotional              string     `json:"long_notional"`
	ShortNotional             string     `json:"short_notional"`
	LongUSD                   string     `json:"long_usd"`
	ShortUSD                  string     `json:"short_usd"`
	InstantDeltaNeutralityFee string     `json:"instant_delta_neutrality_fee_value"`
	DeltaFund                 string     `json:"delta_neutrality_fee_fund"`
	Fees                      Fees       `json:"fees"`
}

// Collateral can be either native or CW20.
type Collateral struct {
	Native *NativeCollateral `json:"native,omitempty"`
	CW20   *CW20Collateral   `json:"cw20,omitempty"`
}

// NativeCollateral represents the native collateral details.
type NativeCollateral struct {
	Denom         string `json:"denom"`
	DecimalPlaces int    `json:"decimal_places"`
}

// CW20Collateral represents the cw20 collateral details.
type CW20Collateral struct {
	Addr          string `json:"addr"`
	DecimalPlaces int    `json:"decimal_places"`
}

// Config represent market configuration
type Config struct {
	TradingFeeNotionalSize        string       `json:"trading_fee_notional_size"`
	TradingFeeCounterCollateral   string       `json:"trading_fee_counter_collateral"`
	CrankExecs                    int64        `json:"crank_execs"`
	MaxLeverage                   string       `json:"max_leverage"`
	FundingRateSensitivity        string       `json:"funding_rate_sensitivity"`
	FundingRateMaxAnnualized      string       `json:"funding_rate_max_annualized"`
	BorrowFeeRateMinAnnualized    string       `json:"borrow_fee_rate_min_annualized"`
	BorrowFeeRateMaxAnnualized    string       `json:"borrow_fee_rate_max_annualized"`
	CarryLeverage                 string       `json:"carry_leverage"`
	MuteEvents                    bool         `json:"mute_events"`
	LiquifundingDelaySeconds      int64        `json:"liquifunding_delay_seconds"`
	ProtocolTax                   string       `json:"protocol_tax"`
	UnstakePeriodSeconds          int64        `json:"unstake_period_seconds"`
	TargetUtilization             string       `json:"target_utilization"`
	BorrowFeeSensitivity          string       `json:"borrow_fee_sensitivity"`
	MaxXLPRewardsMultiplier       string       `json:"max_xlp_rewards_multiplier"`
	MinXLPRewardsMultiplier       string       `json:"min_xlp_rewards_multiplier"`
	DeltaNeutralityFeeSensitivity string       `json:"delta_neutrality_fee_sensitivity"`
	DeltaNeutralityFeeCap         string       `json:"delta_neutrality_fee_cap"`
	DeltaNeutralityFeeTax         string       `json:"delta_neutrality_fee_tax"`
	CrankFeeCharged               string       `json:"crank_fee_charged"`
	CrankFeeSurcharge             string       `json:"crank_fee_surcharge"`
	CrankFeeReward                string       `json:"crank_fee_reward"`
	MinimumDepositUSD             string       `json:"minimum_deposit_usd"`
	LiquifundingDelayFuzzSeconds  int64        `json:"liquifunding_delay_fuzz_seconds"`
	MaxLiquidity                  MaxLiquidity `json:"max_liquidity"`
	DisablePositionNFTExec        bool         `json:"disable_position_nft_exec"`
	LiquidityCooldownSeconds      int64        `json:"liquidity_cooldown_seconds"`
	ExposureMarginRatio           string       `json:"exposure_margin_ratio"`
	ReferralRewardRatio           string       `json:"referral_reward_ratio"`
	SpotPrice                     SpotPrice    `json:"spot_price"`
	PriceUpdateTooOldSeconds      int          `json:"price_update_too_old_seconds"`
	UnpendLimit                   int          `json:"unpend_limit"`
	LimitOrderFee                 string       `json:"limit_order_fee"`
	StalenessSeconds              int          `json:"staleness_seconds"`
}

// Unlimited represents an empty struct for the unlimited field
type Unlimited struct{}

// MaxLiquidity represents the max liquidity configuration
type MaxLiquidity struct {
	Unlimited Unlimited `json:"unlimited"`
}

// SpotPrice represents the spot price details
type SpotPrice struct {
	Oracle Oracle `json:"oracle"`
}

// Oracle represents the oracle configuration
type Oracle struct {
	Pyth            Pyth    `json:"pyth"`
	Stride          Stride  `json:"stride"`
	Feeds           []Feeds `json:"feeds"`
	FeedsUSD        []Feeds `json:"feeds_usd"`
	VolatileDiffSec *string `json:"volatile_diff_seconds"`
}

// Pyth represents the Pyth oracle configuration
type Pyth struct {
	ContractAddress string `json:"contract_address"`
	Network         string `json:"network"`
}

// Stride represents the Stride oracle configuration
type Stride struct {
	ContractAddress string `json:"contract_address"`
}

// Feeds represents the feeds configuration
type Feeds struct {
	Data     FeedsData `json:"data"`
	Inverted bool      `json:"in"`
}

// FeedsData represents the feeds data configuration
type FeedsData struct {
	Pyth PythData `json:"pyth"`
}

// PythData represents the Pyth data for feeds
type PythData struct {
	ID                  string `json:"id"`
	AgeToleranceSeconds int    `json:"age_tolerance_seconds"`
}

// Liquidity represents liquidity details in the response
type Liquidity struct {
	Locked   string `json:"locked"`
	Unlocked string `json:"unlocked"`
	TotalLP  string `json:"total_lp"`
	TotalXLP string `json:"total_xlp"`
}

// Fees represents the fees details in the response
type Fees struct {
	Wallets  string `json:"wallets"`
	Protocol string `json:"protocol"`
	Crank    string `json:"crank"`
	Referral string `json:"referral"`
}

// PositionsResponse represents the response structure for querying positions.
type PositionsResponse struct {
	Positions    []Position `json:"positions"`
	PendingClose []Position `json:"pending_close"`
	Closed       []Position `json:"closed"`
}

// Position represents an individual open, closed, or pending-close position.
type Position struct {
	Owner                        string            `json:"owner"`
	ID                           string            `json:"id"`
	DirectionToBase              string            `json:"direction_to_base"`
	Leverage                     string            `json:"leverage"`
	CounterLeverage              string            `json:"counter_leverage"`
	CreatedAt                    time.UnixNanoTime `json:"created_at"`
	PricePointCreatedAt          time.UnixNanoTime `json:"price_point_created_at"`
	LiquifundedAt                time.UnixNanoTime `json:"liquifunded_at"`
	TradingFeeCollateral         string            `json:"trading_fee_collateral"`
	TradingFeeUSD                string            `json:"trading_fee_usd"`
	FundingFeeCollateral         string            `json:"funding_fee_collateral"`
	FundingFeeUSD                string            `json:"funding_fee_usd"`
	BorrowFeeCollateral          string            `json:"borrow_fee_collateral"`
	BorrowFeeUSD                 string            `json:"borrow_fee_usd"`
	CrankFeeCollateral           string            `json:"crank_fee_collateral"`
	CrankFeeUSD                  string            `json:"crank_fee_usd"`
	DeltaNeutralityFeeCollateral string            `json:"delta_neutrality_fee_collateral"`
	DeltaNeutralityFeeUSD        string            `json:"delta_neutrality_fee_usd"`
	DepositCollateral            string            `json:"deposit_collateral"`
	DepositCollateralUSD         string            `json:"deposit_collateral_usd"`
	ActiveCollateral             string            `json:"active_collateral"`
	ActiveCollateralUSD          string            `json:"active_collateral_usd"`
	CounterCollateral            string            `json:"counter_collateral"`
	PNLCollateral                string            `json:"pnl_collateral"`
	PNLUSD                       string            `json:"pnl_usd"`
	DNFOnCloseCollateral         string            `json:"dnf_on_close_collateral"`
	NotionalSize                 string            `json:"notional_size"`
	NotionalSizeInCollateral     string            `json:"notional_size_in_collateral"`
	PositionSizeBase             string            `json:"position_size_base"`
	PositionSizeUSD              string            `json:"position_size_usd"`
	LiquidationPriceBase         string            `json:"liquidation_price_base"`
	LiquidationMargin            LiquidationMargin `json:"liquidation_margin"`
	MaxGainsInQuote              *string           `json:"max_gains_in_quote"`
	EntryPriceBase               string            `json:"entry_price_base"`
	NextLiquifunding             string            `json:"next_liquifunding"`
	StopLossOverride             *string           `json:"stop_loss_override"`
	TakeProfitOverride           *string           `json:"take_profit_override"`
	TakeProfitPriceBase          string            `json:"take_profit_price_base"`
}

// LiquidationMargin represents the liquidation margin structure in the response.
type LiquidationMargin struct {
	Borrow          string `json:"borrow"`
	Funding         string `json:"funding"`
	DeltaNeutrality string `json:"delta_neutrality"`
	Crank           string `json:"crank"`
	Exposure        string `json:"exposure"`
}

// PositionAction defines a single position action entry
type PositionAction struct {
	ID                 string  `json:"id"`
	Kind               string  `json:"kind"`
	Timestamp          string  `json:"timestamp"`
	PriceTimestamp     string  `json:"price_timestamp"`
	Collateral         string  `json:"collateral"`
	TransferCollateral string  `json:"transfer_collateral"`
	Leverage           *string `json:"leverage"`
	MaxGains           *string `json:"max_gains"`
	TradeFee           *string `json:"trade_fee"`
	DeltaNeutralityFee *string `json:"delta_neutrality_fee"`
	OldOwner           *string `json:"old_owner"`
	NewOwner           *string `json:"new_owner"`
	TakeProfitOverride *string `json:"take_profit_override"`
	StopLossOverride   *string `json:"stop_loss_override"`
}

// PositionActionHistoryResponse defines the response structure for position action history
type PositionActionHistoryResponse struct {
	Actions        []PositionAction `json:"actions"`
	NextStartAfter *string          `json:"next_start_after"`
}

// NotFoundMarker is used to indicate the absence of a deferred exec ID.
type NotFoundMarker struct{}

// GetDeferredExecIDResponse represents the response from querying a deferred exec ID.
type GetDeferredExecIDResponse struct {
	Found    *DeferredExecFound `json:"found,omitempty"`
	NotFound *NotFoundMarker    `json:"not_found,omitempty"`
}

// DeferredExecFound represents the found deferred execution item.
type DeferredExecFound struct {
	Item DeferredExecItem `json:"item"`
}

// DeferredExecItem represents the actual deferred execution item details.
type DeferredExecItem struct {
	ID      string                  `json:"id"`
	Created string                  `json:"created"`
	Status  json.RawMessage         `json:"status"` // Store raw JSON to dynamically parse later
	Owner   string                  `json:"owner"`
	Item    DeferredExecItemDetails `json:"item"`
}

// DeferredExecStatus represents the execution status.
type DeferredExecStatus struct {
	Success *DeferredExecSuccess `json:"success,omitempty"`
}

// DeferredExecSuccess represents a successfully executed deferred execution.
type DeferredExecSuccess struct {
	Target   DeferredExecTarget `json:"target"`
	Executed string             `json:"executed"`
}

// DeferredExecTarget represents the execution target.
type DeferredExecTarget struct {
	Position string `json:"position"`
}

// DeferredExecItemDetails represents the open position details.
type DeferredExecItemDetails struct {
	OpenPosition OpenPositionDetails `json:"open_position"`
}

// OpenPositionDetails represents the specifics of the open position.
type OpenPositionDetails struct {
	SlippageAssert   SlippageAssertDetails `json:"slippage_assert"`
	Leverage         string                `json:"leverage"`
	Direction        string                `json:"direction"`
	MaxGains         *string               `json:"max_gains,omitempty"`
	StopLossOverride *string               `json:"stop_loss_override,omitempty"`
	TakeProfit       *string               `json:"take_profit,omitempty"`
	Amount           string                `json:"amount"`
	CrankFee         string                `json:"crank_fee"`
	CrankFeeUSD      string                `json:"crank_fee_usd"`
}

// SlippageAssertDetails represents slippage assertion details.
type SlippageAssertDetails struct {
	Price     string `json:"price"`
	Tolerance string `json:"tolerance"`
}
