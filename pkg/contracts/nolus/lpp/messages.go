package lpp

import (
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BurnRequest is the message for burning nLPNs to withdraw funds
type BurnRequest struct {
	Amount uint64 `json:"amount"`
}

// BuildDepositMsg constructs a Deposit payload and returns a MsgExecuteContract
func BuildDepositMsg(sender, contractAddress string, funds sdk.Coins) (*wasmtypes.MsgExecuteContract, error) {
	// Construct the Deposit message
	depositMsg := map[string]interface{}{
		"deposit": struct{}{},
	}

	// Convert the Deposit message to JSON bytes
	executeMsgBytes, err := json.Marshal(depositMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Deposit message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    funds,
	}

	return msgExecuteContract, nil
}

// BuildBurnMsg constructs a Burn payload and returns a MsgExecuteContract
func BuildBurnMsg(sender, contractAddress string, amount uint64) (*wasmtypes.MsgExecuteContract, error) {
	// Construct the Burn message
	burnMsg := map[string]interface{}{
		"burn": BurnRequest{
			Amount: amount,
		},
	}

	// Convert the Burn message to JSON bytes
	executeMsgBytes, err := json.Marshal(burnMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Burn message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    sdk.Coins{},
	}

	return msgExecuteContract, nil
}
