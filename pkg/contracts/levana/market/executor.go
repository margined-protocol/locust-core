package market

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// ExecuteMsgGenerator handles execution for both CW20-based and native transactions
// ExecuteMsgGenerator handles execution for CW20-based and native transactions dynamically
type ExecuteMsgGenerator struct{}

// NewMsgGenerator creates a new instance of ExecuteMsgGenerator
func NewMsgGenerator() *ExecuteMsgGenerator {
	return &ExecuteMsgGenerator{}
}

// Execute handles execution, supporting both CW20 and native funds
func (p *ExecuteMsgGenerator) Execute(
	sender string,
	contractAddress string, // Contract address is now passed dynamically
	action string,
	payload interface{},
	funds *sdktypes.Coins, // Optional native funds
	useCW20Contract *string, // Optional override for CW20 contract address
) (*wasmdtypes.MsgExecuteContract, error) {
	// Marshal the inner execution message
	innerMsg := map[string]interface{}{action: payload}
	innerMsgBytes, err := json.Marshal(innerMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s message: %w", action, err)
	}

	// If using CW20, wrap in a "send" message
	if useCW20Contract != nil && funds != nil && !funds.IsZero() {
		// Extract amount from the first coin
		amount := (*funds)[0].Amount.String()

		// Base64 encode the message
		encodedMsg := base64.StdEncoding.EncodeToString(innerMsgBytes)

		// Construct the CW20 send message
		sendMsg := map[string]interface{}{
			"send": map[string]interface{}{
				"amount":   amount,
				"contract": contractAddress, // Target market contract
				"msg":      encodedMsg,
			},
		}
		sendMsgBytes, err := json.Marshal(sendMsg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal CW20 send message: %w", err)
		}

		return &wasmdtypes.MsgExecuteContract{
			Sender:   sender,
			Contract: *useCW20Contract, // Send via CW20 contract
			Msg:      sendMsgBytes,
			Funds:    sdktypes.Coins{}, // No native funds needed
		}, nil
	}

	// Otherwise, execute directly on the market contract with native funds
	if funds == nil {
		funds = &sdktypes.Coins{} // Ensure empty funds are handled correctly
	}

	return &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress, // Execute directly
		Msg:      innerMsgBytes,
		Funds:    *funds, // Include native funds if applicable
	}, nil
}
