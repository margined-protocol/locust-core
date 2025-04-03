package strategy

import (
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// CreateStrategyWithdrawMsg constructs a MsgExecuteContract to withdraw funds from the strategy contract.
func CreateStrategyWithdrawMsg(sender, contractAddress string, coins []sdktypes.Coin) (*wasmdtypes.MsgExecuteContract, error) {
	if len(coins) == 0 {
		return nil, fmt.Errorf("coins must not be empty")
	}

	withdrawMsg := &WithdrawMessage{
		Withdraw: WithdrawDetails{
			TokensToWithdraw: coins,
		},
	}

	withdrawMsgBytes, err := json.Marshal(withdrawMsg)
	if err != nil {
		return nil, err
	}

	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      withdrawMsgBytes,
		Funds:    sdktypes.Coins{},
	}

	return msgExecuteContract, nil
}

// CreateStrategyRepayMsg constructs a MsgExecuteContract to repay funds into the strategy contract.
func CreateStrategyRepayMsg(sender, contractAddress, cycleProfit string, coins []sdktypes.Coin) (*wasmdtypes.MsgExecuteContract, error) {
	repayMsg := &RepayMessage{
		Repay: RepayDetails{
			TokensToRepay: coins,
		},
	}

	if cycleProfit != "" {
		repayMsg.Repay.CycleProfit = cycleProfit
	}

	repayMsgBytes, err := json.Marshal(repayMsg)
	if err != nil {
		return nil, err
	}

	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      repayMsgBytes,
		Funds:    sdktypes.Coins{},
	}

	return msgExecuteContract, nil
}

// CreateStrategyRepayQueueMsg constructs a MsgExecuteContract to repay funds from strategy contract to fund
func CreateStrategyRepayQueueMsg(sender, contractAddress, cycleProfit string, coins []sdktypes.Coin) (*wasmdtypes.MsgExecuteContract, error) {
	repayMsg := &RepayQueueMessage{
		Repay: RepayDetails{
			TokensToRepay: coins,
		},
	}

	if cycleProfit != "" {
		repayMsg.Repay.CycleProfit = cycleProfit
	}

	repayMsgBytes, err := json.Marshal(repayMsg)
	if err != nil {
		return nil, err
	}

	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      repayMsgBytes,
		Funds:    sdktypes.Coins{},
	}

	return msgExecuteContract, nil
}
