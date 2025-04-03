package market

import (
	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// LevanaProvider handles contract interactions dynamically
type LevanaProvider struct {
	msgGenerator *ExecuteMsgGenerator
}

// NewLevanaProvider initializes LevanaProvider
func NewLevanaProvider() *LevanaProvider {
	return &LevanaProvider{
		msgGenerator: NewMsgGenerator(),
	}
}

// OpenPosition executes `open_position` on the given contract
func (p *LevanaProvider) OpenPosition(
	sender, contractAddress string,
	cw20Contract *string, // CW20 contract, nil if not CW20
	payload OpenPositionMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "open_position", payload, funds, cw20Contract)
}

// ClosePosition executes `close_position` on the given contract
func (p *LevanaProvider) ClosePosition(
	sender, contractAddress string,
	cw20Contract *string,
	payload ClosePositionMsg,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "close_position", payload, nil, cw20Contract)
}

// UpdateTakeProfitPrice executes `update_position_take_profit_price`
func (p *LevanaProvider) UpdateTakeProfitPrice(
	sender, contractAddress string,
	cw20Contract *string,
	payload UpdatePositionTakeProfitPriceMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "update_position_take_profit_price", payload, funds, cw20Contract)
}

// PlaceLimitOrder executes `place_limit_order`
func (p *LevanaProvider) PlaceLimitOrder(
	sender, contractAddress string,
	cw20Contract *string,
	payload PlaceLimitOrderMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "place_limit_order", payload, funds, cw20Contract)
}

// CancelLimitOrder executes `cancel_limit_order`
func (p *LevanaProvider) CancelLimitOrder(
	sender, contractAddress string,
	cw20Contract *string,
	payload CancelLimitOrderMsg,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "cancel_limit_order", payload, nil, cw20Contract)
}

// AddCollateral executes `update_position_add_collateral_impact_leverage`
func (p *LevanaProvider) AddCollateral(
	sender, contractAddress string,
	cw20Contract *string,
	payload UpdatePositionAddCollateralImpactLeverageMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "update_position_add_collateral_impact_leverage", payload, funds, cw20Contract)
}

// RemoveCollateral executes `update_position_remove_collateral_impact_leverage`
func (p *LevanaProvider) RemoveCollateral(
	sender, contractAddress string,
	cw20Contract *string,
	payload UpdatePositionRemoveCollateralImpactLeverageMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "update_position_remove_collateral_impact_leverage", payload, funds, cw20Contract)
}

// UpdateLeverage executes `update_position_leverage`
func (p *LevanaProvider) UpdateLeverage(
	sender, contractAddress string,
	cw20Contract *string,
	payload UpdatePositionLeverageMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "update_position_leverage", payload, funds, cw20Contract)
}

// UpdateStopLoss executes `update_position_stop_loss_price`
func (p *LevanaProvider) UpdateStopLoss(
	sender, contractAddress string,
	cw20Contract *string,
	payload UpdatePositionStopLossPriceMsg,
	funds *sdktypes.Coins,
) (*wasmdtypes.MsgExecuteContract, error) {
	return p.msgGenerator.Execute(sender, contractAddress, "update_position_stop_loss_price", payload, funds, cw20Contract)
}

// RemoveStopLoss executes `update_position_stop_loss_price` with "remove"
func (p *LevanaProvider) RemoveStopLoss(
	sender, contractAddress string,
	cw20Contract *string,
	positionID string,
) (*wasmdtypes.MsgExecuteContract, error) {
	payload := map[string]interface{}{
		"update_position_stop_loss_price": map[string]interface{}{
			"id":        positionID,
			"stop_loss": "remove",
		},
	}
	return p.msgGenerator.Execute(sender, contractAddress, "update_position_stop_loss_price", payload, nil, cw20Contract)
}
