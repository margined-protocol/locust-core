package levana

type ExecuteMsg struct {
	// Owner                                        *ExecuteOwnerMsg                              `json:"owner,omitempty"`
	Receive                                      *ReceiveMsg                                   `json:"receive,omitempty"`
	OpenPosition                                 *OpenPosition                                 `json:"open_position,omitempty"`
	UpdatePositionAddCollateralImpactLeverage    *UpdatePositionAddCollateralImpactLeverage    `json:"update_position_add_collateral_impact_leverage,omitempty"`
	UpdatePositionAddCollateralImpactSize        *UpdatePositionAddCollateralImpactSize        `json:"update_position_add_collateral_impact_size,omitempty"`
	UpdatePositionRemoveCollateralImpactLeverage *UpdatePositionRemoveCollateralImpactLeverage `json:"update_position_remove_collateral_impact_leverage,omitempty"`
	UpdatePositionRemoveCollateralImpactSize     *UpdatePositionRemoveCollateralImpactSize     `json:"update_position_remove_collateral_impact_size,omitempty"`
	UpdatePositionLeverage                       *UpdatePositionLeverage                       `json:"update_position_leverage,omitempty"`
	UpdatePositionMaxGains                       *UpdatePositionMaxGains                       `json:"update_position_max_gains,omitempty"`
	UpdatePositionTakeProfitPrice                *UpdatePositionTakeProfitPrice                `json:"update_position_take_profit_price,omitempty"`
	UpdatePositionStopLossPrice                  *UpdatePositionStopLossPrice                  `json:"update_position_stop_loss_price,omitempty"`
	SetTriggerOrder                              *SetTriggerOrder                              `json:"set_trigger_order,omitempty"`
	PlaceLimitOrder                              *PlaceLimitOrder                              `json:"place_limit_order,omitempty"`
	CancelLimitOrder                             *CancelLimitOrder                             `json:"cancel_limit_order,omitempty"`
	ClosePosition                                *ClosePosition                                `json:"close_position,omitempty"`
	DepositLiquidity                             *DepositLiquidity                             `json:"deposit_liquidity,omitempty"`
	ReinvestYield                                *ReinvestYield                                `json:"reinvest_yield,omitempty"`
	WithdrawLiquidity                            *WithdrawLiquidity                            `json:"withdraw_liquidity,omitempty"`
	ClaimYield                                   *ClaimYield                                   `json:"claim_yield,omitempty"`
	StakeLp                                      *StakeLp                                      `json:"stake_lp,omitempty"`
	UnstakeXlp                                   *UnstakeXlp                                   `json:"unstake_xlp,omitempty"`
	StopUnstakingXlp                             *StopUnstakingXlp                             `json:"stop_unstaking_xlp,omitempty"`
	CollectUnstakedLp                            *CollectUnstakedLp                            `json:"collect_unstaked_lp,omitempty"`
	Crank                                        *Crank                                        `json:"crank,omitempty"`
	NftProxy                                     *NftProxy                                     `json:"nft_proxy,omitempty"`
	LiquidityTokenProxy                          *LiquidityTokenProxy                          `json:"liquidity_token_proxy,omitempty"`
	TransferDaoFees                              *TransferDaoFees                              `json:"transfer_dao_fees,omitempty"`
	CloseAllPositions                            *CloseAllPositions                            `json:"close_all_positions,omitempty"`
	ProvideCrankFunds                            *ProvideCrankFunds                            `json:"provide_crank_funds,omitempty"`
	SetManualPrice                               *SetManualPrice                               `json:"set_manual_price,omitempty"`
	PerformDeferredExec                          *PerformDeferredExec                          `json:"perform_deferred_exec,omitempty"`
}

// type ExecuteOwnerMsg struct {
// 	ConfigUpdate *ConfigUpdate `json:"config_update"`
// }

type ReceiveMsg struct {
	Sender string `json:"sender"`
	Amount string `json:"amount"`
	Msg    string `json:"msg"`
}

type OpenPosition struct {
	SlippageAssert   *SlippageAssert `json:"slippage_assert,omitempty"`
	Leverage         string          `json:"leverage"`
	Direction        string          `json:"direction"`
	MaxGains         *string         `json:"max_gains,omitempty"`
	StopLossOverride *string         `json:"stop_loss_override,omitempty"`
	TakeProfit       *string         `json:"take_profit,omitempty"`
}

type UpdatePositionAddCollateralImpactLeverage struct {
	ID string `json:"id"`
}

type UpdatePositionAddCollateralImpactSize struct {
	ID             string          `json:"id"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

type UpdatePositionRemoveCollateralImpactLeverage struct {
	ID     string `json:"id"`
	Amount string `json:"amount"`
}

type UpdatePositionRemoveCollateralImpactSize struct {
	ID             string          `json:"id"`
	Amount         string          `json:"amount"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

type UpdatePositionLeverage struct {
	ID             string          `json:"id"`
	Leverage       string          `json:"leverage"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

type UpdatePositionMaxGains struct {
	ID       string `json:"id"`
	MaxGains string `json:"max_gains"`
}

type UpdatePositionTakeProfitPrice struct {
	ID    string `json:"id"`
	Price string `json:"price"`
}

type UpdatePositionStopLossPrice struct {
	ID       string   `json:"id"`
	StopLoss StopLoss `json:"stop_loss"`
}

type SetTriggerOrder struct {
	ID               string  `json:"id"`
	StopLossOverride *string `json:"stop_loss_override,omitempty"`
	TakeProfit       *string `json:"take_profit,omitempty"`
}

type PlaceLimitOrder struct {
	TriggerPrice     string  `json:"trigger_price"`
	Leverage         string  `json:"leverage"`
	Direction        string  `json:"direction"`
	MaxGains         *string `json:"max_gains,omitempty"`
	StopLossOverride *string `json:"stop_loss_override,omitempty"`
	TakeProfit       *string `json:"take_profit,omitempty"`
}

type CancelLimitOrder struct {
	OrderID string `json:"order_id"`
}

type ClosePosition struct {
	ID             string          `json:"id"`
	SlippageAssert *SlippageAssert `json:"slippage_assert,omitempty"`
}

type DepositLiquidity struct {
	StakeToXlp bool `json:"stake_to_xlp"`
}

type ReinvestYield struct {
	StakeToXlp bool    `json:"stake_to_xlp"`
	Amount     *string `json:"amount,omitempty"`
}

type WithdrawLiquidity struct {
	LpAmount *string `json:"lp_amount,omitempty"`
}

type ClaimYield struct{}

type StakeLp struct {
	Amount *string `json:"amount,omitempty"`
}

type UnstakeXlp struct {
	Amount *string `json:"amount,omitempty"`
}

type StopUnstakingXlp struct{}

type CollectUnstakedLp struct{}

type Crank struct {
	Execs   *uint32 `json:"execs,omitempty"`
	Rewards *string `json:"rewards,omitempty"`
}

type NftProxy struct {
	Sender string `json:"sender"`
	Msg    string `json:"msg"`
}

type LiquidityTokenProxy struct {
	Sender string `json:"sender"`
	Kind   string `json:"kind"`
	Msg    string `json:"msg"`
}

type TransferDaoFees struct{}

type CloseAllPositions struct{}

type ProvideCrankFunds struct{}

type SetManualPrice struct {
	Price    string `json:"price"`
	PriceUsd string `json:"price_usd"`
}

type PerformDeferredExec struct {
	ID                  string `json:"id"`
	PricePointTimestamp string `json:"price_point_timestamp"`
}

// Supporting Types
type SlippageAssert struct {
	Price     string `json:"price"`
	Tolerance string `json:"tolerance"`
}

type StopLoss struct {
	Price string `json:"price"`
}
