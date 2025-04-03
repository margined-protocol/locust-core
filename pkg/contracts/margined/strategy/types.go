package strategy

import sdktypes "github.com/cosmos/cosmos-sdk/types"

// WithdrawMessage represents the payload structure for the contract call.
type WithdrawMessage struct {
	Withdraw WithdrawDetails `json:"withdraw"`
}

// WithdrawDetails holds the details for the swap action.
type WithdrawDetails struct {
	TokensToWithdraw []sdktypes.Coin `json:"tokens_to_withdraw"`
}

// RepayMessage represents the payload structure for the contract call.
type RepayMessage struct {
	Repay RepayDetails `json:"repay"`
}

// RepayMessage represents the payload structure for the contract call.
type RepayQueueMessage struct {
	Repay RepayDetails `json:"repay_queue"`
}

// RepayDetails holds the details for the swap action.
type RepayDetails struct {
	TokensToRepay []sdktypes.Coin `json:"tokens_to_repay"`
	CycleProfit   string          `json:"cycle_profit,omitempty"`
	Limit         uint64          `json:"limit,omitempty"`
}
