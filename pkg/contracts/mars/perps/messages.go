package perps

import (
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types"
)

// BuildExecuteOrderMsg constructs an ExecuteOrder payload and returns a MsgExecuteContract
func BuildExecuteOrderMsg(sender, contractAddress, accountID, denom string, size string, reduceOnly *bool) (*wasmdtypes.MsgExecuteContract, error) {
	// Construct the ExecuteOrder message
	executeOrderMsg := map[string]interface{}{
		"execute_order": map[string]interface{}{
			"account_id": accountID,
			"denom":      denom,
			"size":       size,
		},
	}

	// Optionally include the reduce_only field if it is not nil
	if reduceOnly != nil {
		executeOrderMsg["execute_order"].(map[string]interface{})["reduce_only"] = *reduceOnly
	}

	// Convert the ExecuteOrder message to JSON bytes
	executeMsgBytes, err := json.Marshal(executeOrderMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ExecuteOrder message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    types.Coins{}, // Add funds if needed
	}

	return msgExecuteContract, nil
}
