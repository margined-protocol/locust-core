package creditmanager

import (
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/types"
)

// BuildCreateCreditAccountMsg constructs a CreateCreditAccount payload and returns a MsgExecuteContract
func BuildCreateCreditAccountMsg(sender, contractAddress string, accountKind AccountKind) (*wasmdtypes.MsgExecuteContract, error) {
	// Construct the CreateCreditAccount message
	createCreditAccountMsg := map[string]interface{}{
		"create_credit_account": accountKind,
	}

	// Convert the CreateCreditAccount message to JSON bytes
	executeMsgBytes, err := json.Marshal(createCreditAccountMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CreateCreditAccount message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    types.Coins{},
	}

	return msgExecuteContract, nil
}

// BuildUpdateCreditAccountMsg constructs an UpdateCreditAccount payload and returns a MsgExecuteContract
func BuildUpdateCreditAccountMsg(sender, contractAddress string, accountID *string, actions []Action, funds types.Coins) (*wasmdtypes.MsgExecuteContract, error) {
	// Construct the UpdateCreditAccount message
	updateCreditAccountMsg := map[string]interface{}{
		"update_credit_account": map[string]interface{}{
			"account_id": accountID,
			"actions":    actions,
		},
	}

	// Convert the UpdateCreditAccount message to JSON bytes
	executeMsgBytes, err := json.Marshal(updateCreditAccountMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal UpdateCreditAccount message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    funds,
	}

	return msgExecuteContract, nil
}

// BuildRepayFromWalletMsg constructs a RepayFromWallet payload and returns a MsgExecuteContract
func BuildRepayFromWalletMsg(sender, contractAddress, accountID string, funds types.Coins) (*wasmdtypes.MsgExecuteContract, error) {
	// Construct the RepayFromWallet message
	repayFromWalletMsg := map[string]interface{}{
		"repay_from_wallet": map[string]interface{}{
			"account_id": accountID,
		},
	}

	// Convert the RepayFromWallet message to JSON bytes
	executeMsgBytes, err := json.Marshal(repayFromWalletMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RepayFromWallet message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    funds,
	}

	return msgExecuteContract, nil
}

// BuildExecuteTriggerOrderMsg constructs an ExecuteTriggerOrder payload and returns a MsgExecuteContract
func BuildExecuteTriggerOrderMsg(sender, contractAddress, accountID, triggerOrderID string) (*wasmdtypes.MsgExecuteContract, error) {
	// Construct the ExecuteTriggerOrder message
	executeTriggerOrderMsg := map[string]interface{}{
		"execute_trigger_order": map[string]interface{}{
			"account_id":       accountID,
			"trigger_order_id": triggerOrderID,
		},
	}

	// Convert the ExecuteTriggerOrder message to JSON bytes
	executeMsgBytes, err := json.Marshal(executeTriggerOrderMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ExecuteTriggerOrder message: %w", err)
	}

	// Construct the MsgExecuteContract message
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      executeMsgBytes,
		Funds:    types.Coins{},
	}

	return msgExecuteContract, nil
}
