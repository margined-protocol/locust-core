package drop

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types"
)

// Cw721ReceiveMsg represents the receive message structure for a CW721 NFT
type Cw721ReceiveMsg struct {
	Msg     string `json:"msg"`
	Sender  string `json:"sender"`
	TokenID string `json:"token_id"`
}

// BuildSendNftMsg constructs a SendNft payload and returns a MsgExecuteContract
func BuildSendNftMsg(sender, withdrawalManagerAddress, withdrawalVoucherAddress, tokenID string) (*wasmdtypes.MsgExecuteContract, error) {
	withdrawMsg := `{"withdraw":{}}`
	encodedWithdrawMsg := base64.StdEncoding.EncodeToString([]byte(withdrawMsg))

	// Construct the SendNft message with the base64-encoded withdraw message
	sendNftMsg := map[string]interface{}{
		"contract": withdrawalManagerAddress,
		"token_id": tokenID,
		"msg":      encodedWithdrawMsg,
	}

	// Convert the SendNft message to JSON bytes
	sendNftMsgBytes, err := json.Marshal(sendNftMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SendNft message: %w", err)
	}

	// Create the outer message structure for MsgExecuteContract
	executeMsg := map[string]json.RawMessage{
		"send_nft": sendNftMsgBytes,
	}

	// Convert the outer message to JSON
	executeMsgBytes, err := json.Marshal(executeMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execute message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: withdrawalVoucherAddress,
		Msg:      executeMsgBytes,
		Funds:    types.Coins{}, // No funds are sent with the message
	}

	return msgExecuteContract, nil
}

// BuildCw721ReceiveMsg constructs a Cw721ReceiveMsg payload and returns a MsgExecuteContract
func BuildCw721ReceiveMsg(sender, contractAddress, tokenID string) (*wasmdtypes.MsgExecuteContract, error) {
	// Create the receive message
	withdrawMsg := map[string]interface{}{
		"withdraw": map[string]interface{}{
			"receiver": sender,
		},
	}
	// Marshal the inner message to JSON bytes
	withdrawMsgBytes, err := json.Marshal(withdrawMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal withdraw message: %w", err)
	}

	// Base64 encode the withdraw message
	encodedWithdrawMsg := base64.StdEncoding.EncodeToString(withdrawMsgBytes)

	receiveMsg := Cw721ReceiveMsg{
		Msg:     encodedWithdrawMsg,
		Sender:  sender,
		TokenID: tokenID,
	}

	// Convert the receive message to JSON
	receiveMsgBytes, err := json.Marshal(receiveMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal receive message: %w", err)
	}

	// Create the outer message structure
	executeMsg := map[string]json.RawMessage{
		"receive_nft": receiveMsgBytes,
	}

	// Convert the outer message to JSON
	executeMsgBytes, err := json.Marshal(executeMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execute message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    types.Coins{}, // No funds are sent with the message
	}

	return msgExecuteContract, nil
}

// CreateUnbondMsg constructs the unbond message with funds and returns a MsgExecuteContract
func CreateUnbondMsg(sender, contractAddress string, funds types.Coin) (*wasmdtypes.MsgExecuteContract, error) {
	unbondMsg := map[string]interface{}{
		"unbond": map[string]interface{}{},
	}

	// Convert the unbond message to JSON
	unbondMsgBytes, err := json.Marshal(unbondMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unbond message: %w", err)
	}

	// Construct the MsgExecuteContract message with funds
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      unbondMsgBytes,
		Funds:    types.Coins{funds}, // Funds to be sent with the message
	}

	return msgExecuteContract, nil
}
