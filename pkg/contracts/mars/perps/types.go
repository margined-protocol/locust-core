package perps

import (
	"math/big"

	"github.com/cosmos/cosmos-sdk/types"
)

// MarketStateResponse represents the response for market state queries.
type MarketStateResponse struct {
	Denom       string      `json:"denom"`
	MarketState MarketState `json:",inline"`
}

// VaultPositionResponse represents the response for vault positions.
type VaultPositionResponse struct {
	Denom   string        `json:"denom"`
	Deposit VaultDeposit  `json:"deposit"`
	Unlocks []VaultUnlock `json:"unlocks"`
}

// VaultDeposit represents details of a vault deposit.
type VaultDeposit struct {
	Shares *big.Int `json:"shares"`
	Amount *big.Int `json:"amount"`
}

// VaultUnlock represents details of a vault unlock.
type VaultUnlock struct {
	CreatedAt   uint64   `json:"created_at"`
	CooldownEnd uint64   `json:"cooldown_end"`
	Shares      *big.Int `json:"shares"`
	Amount      *big.Int `json:"amount"`
}

// PositionResponse represents the response for a single position.
type PositionResponse struct {
	AccountID string        `json:"account_id"`
	Position  *PerpPosition `json:"position,omitempty"`
}

// PositionsByAccountResponse represents the response for all positions by account.
type PositionsByAccountResponse struct {
	AccountID string         `json:"account_id"`
	Positions []PerpPosition `json:"positions"`
}

// TradingFee represents the trading fee details.
type TradingFee struct {
	Rate string     `json:"rate"`
	Fee  types.Coin `json:"fee"`
}

// PositionFeesResponse represents the fees associated with modifying a position.
type PositionFeesResponse struct {
	BaseDenom        string   `json:"base_denom"`
	OpeningFee       *big.Int `json:"opening_fee"`
	ClosingFee       *big.Int `json:"closing_fee"`
	OpeningExecPrice *string  `json:"opening_exec_price,omitempty"`
	ClosingExecPrice *string  `json:"closing_exec_price,omitempty"`
}

// MarketResponse represents the market state for a single denom.
type MarketResponse struct {
	Denom              string   `json:"denom"`                // Denomination of the asset
	Enabled            bool     `json:"enabled"`              // Enabled status of the market
	LongOI             *big.Int `json:"long_oi"`              // Total LONG open interest in utokens
	LongOIValue        *big.Int `json:"long_oi_value"`        // Total LONG open interest in oracle base currency (uusd)
	ShortOI            *big.Int `json:"short_oi"`             // Total SHORT open interest in utokens
	ShortOIValue       *big.Int `json:"short_oi_value"`       // Total SHORT open interest in oracle base currency (uusd)
	CurrentFundingRate string   `json:"current_funding_rate"` // Current funding rate
}

// MarketState represents the flattened market state.
type MarketState struct {
	Enabled                     bool     `json:"enabled"`                        // Whether the denom is enabled for trading
	LongOI                      *big.Int `json:"long_oi"`                        // Total LONG open interest
	ShortOI                     *big.Int `json:"short_oi"`                       // Total SHORT open interest
	TotalEntryCost              *big.Int `json:"total_entry_cost"`               // Accumulated entry cost
	TotalEntryFunding           *big.Int `json:"total_entry_funding"`            // Accumulated entry funding
	TotalSquaredPositions       *big.Int `json:"total_squared_positions"`        // Accumulated squared positions
	TotalAbsMultipliedPositions *big.Int `json:"total_abs_multiplied_positions"` // Accumulated absolute multiplied positions
	CashFlow                    CashFlow `json:"cash_flow"`                      // Actual amount of money (realized payments)
	Funding                     Funding  `json:"funding"`                        // Funding parameters for this denom
	LastUpdated                 uint64   `json:"last_updated"`                   // Last time this denom was updated
}

// PerpPosition represents a perpetual position.
type PerpPosition struct {
	Denom            string     `json:"denom"`
	BaseDenom        string     `json:"base_denom"`
	Size             *string    `json:"size"`
	EntryPrice       string     `json:"entry_price"`        // Decimal represented as a string
	CurrentPrice     string     `json:"current_price"`      // Decimal represented as a string
	EntryExecPrice   string     `json:"entry_exec_price"`   // Decimal represented as a string
	CurrentExecPrice string     `json:"current_exec_price"` // Decimal represented as a string
	UnrealizedPnl    PnlAmounts `json:"unrealized_pnl"`
	RealizedPnl      PnlAmounts `json:"realized_pnl"`
}

// CashFlow represents the actual amount of money, including only realized payments.
type CashFlow struct {
	Amount *big.Int `json:"amount"`
}

// Funding represents the funding parameters for a specific denom.
type Funding struct {
	Rate        string `json:"rate"`
	Accumulated string `json:"accumulated"`
}

// PnlAmounts represents amounts denominated in the Perp Vault base denom (uusdc).
type PnlAmounts struct {
	PricePnl       *string `json:"price_pnl"`       // Price PnL
	AccruedFunding *string `json:"accrued_funding"` // Accrued funding
	OpeningFee     *string `json:"opening_fee"`     // Opening fee
	ClosingFee     *string `json:"closing_fee"`     // Closing fee

	// PnL: price PnL + accrued funding + opening fee + closing fee
	Pnl *string `json:"pnl"`
}

// MarketStateRequest is the request type for querying the state of a specific market.
type MarketStateRequest struct {
	Denom string `json:"denom"`
}

// MarketRequest is the request type for querying a single market.
type MarketRequest struct {
	Denom string `json:"denom"`
}

// MarketsRequest is the request type for querying markets with pagination.
type MarketsRequest struct {
	StartAfter *string `json:"start_after,omitempty"`
	Limit      *uint32 `json:"limit,omitempty"`
}

// PositionRequest is the request type for querying a single perp position by account and denom.
type PositionRequest struct {
	AccountID  string   `json:"account_id"`
	Denom      string   `json:"denom"`
	OrderSize  *big.Int `json:"order_size,omitempty"`  // Optional
	ReduceOnly *bool    `json:"reduce_only,omitempty"` // Optional
}

// PositionsRequest is the request type for listing positions of all accounts and denoms.
type PositionsRequest struct {
	StartAfter *struct {
		AccountID string `json:"account_id"`
		Denom     string `json:"denom"`
	} `json:"start_after,omitempty"`
	Limit *uint32 `json:"limit,omitempty"`
}

// PositionsByAccountRequest is the request type for listing positions of all denoms
// that belong to a specific credit account.
type PositionsByAccountRequest struct {
	AccountID string      `json:"account_id"`
	Action    *ActionKind `json:"action,omitempty"`
}

// RealizedPnlRequest is the request type for querying realized PnL amounts
// for a specific account and market.
type RealizedPnlRequest struct {
	AccountID string `json:"account_id"`
	Denom     string `json:"denom"`
}

// MarketAccountingRequest is the request type for querying the accounting details for a specific market.
type MarketAccountingRequest struct {
	Denom string `json:"denom"`
}

// OpeningFeeRequest is the request type for querying the opening fee for a given market and position size.
type OpeningFeeRequest struct {
	Denom string   `json:"denom"`
	Size  *big.Int `json:"size"`
}

// PositionFeesRequest is the request type for querying the fees associated with modifying a specific position.
type PositionFeesRequest struct {
	AccountID string   `json:"account_id"`
	Denom     string   `json:"denom"`
	NewSize   *big.Int `json:"new_size"`
}

// ActionKind represents the differentiator for the action being performed.
type ActionKind string

const (
	// Default action kind
	ActionKindDefault ActionKind = "Default"

	// Liquidation action kind
	ActionKindLiquidation ActionKind = "Liquidation"
)
