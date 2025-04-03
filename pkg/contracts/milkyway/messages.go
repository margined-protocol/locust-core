package milkyway

import (
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

func CreateWithdrawMessage(sender, contractAddress string, batchID uint64) (sdktypes.Msg, error) {
	msg := WithdrawMessage{
		Withdraw: &WithdrawDetails{
			BatchID: batchID,
		},
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal milkyway message: %w", err)
	}

	// Construct the MsgExecuteContract with the generated message and funds
	msgExecuteContract := wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      msgBytes,
		Funds:    []sdktypes.Coin{},
	}

	return &msgExecuteContract, nil
}
