package pyth

import (
	"encoding/json"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CreateUpdatePriceFeedsMsg constructs and returns the MsgExecuteContract to execute the "updatePriceFeeds" method on the contract
func CreateUpdatePriceFeedsMsg(contractAddress string, senderAddress string, base64Data string, funds sdk.Coins) (*wasmdtypes.MsgExecuteContract, error) {
	executeMsg := UpdatePriceFeedsMsg{
		UpdatePriceFeeds: UpdatePriceFeeds{
			Data: []string{base64Data},
		},
	}

	msgBytes, err := json.Marshal(executeMsg)
	if err != nil {
		return nil, err
	}

	msg := wasmdtypes.MsgExecuteContract{
		Sender:   senderAddress,
		Contract: contractAddress,
		Msg:      msgBytes,
		Funds:    funds,
	}

	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	return &msg, nil
}
