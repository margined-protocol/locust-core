package lsd

import (
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// VaultenatorExtensionExecuteMsg represents the innermost data structure.
type VaultenatorExtensionExecuteMsg struct {
	WithdrawAndSwap *WithdrawAndSwap `json:"withdraw_and_swap,omitempty"`
	Repay           *Repay           `json:"repay,omitempty"`
	Crank           *Crank           `json:"crank,omitempty"`
}

// WithdrawAndSwap represents the fields of the WithdrawAndSwap variant.
type WithdrawAndSwap struct {
	WithdrawAmount string `json:"withdraw_amount"`
	MinAmountOut   string `json:"min_amount_out"`
	PoolID         uint64 `json:"pool_id"`
	Denom          string `json:"denom"`
}

// ExtensionExecuteMsg represents the Vaultenator variant.
type ExtensionExecuteMsg struct {
	Vaultenator *VaultenatorExtensionExecuteMsg `json:"vaultenator,omitempty"`
}

// ExecuteMsg represents the top-level VaultExtension variant.
type ExecuteMsg struct {
	VaultExtension *ExtensionExecuteMsg `json:"vault_extension,omitempty"`
}

type (
	Repay struct{}
	Crank struct{}
)

func CreateCrankMsg(sender, contract string) sdk.Msg {
	// Populate the struct with your variables
	msgStruct := ExecuteMsg{
		VaultExtension: &ExtensionExecuteMsg{
			Vaultenator: &VaultenatorExtensionExecuteMsg{
				Crank: &Crank{},
			},
		},
	}

	// Marshal the struct into JSON
	msgBytes, err := json.Marshal(msgStruct)
	if err != nil {
		panic(err)
	}
	// Generate the swap message
	msg := wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contract,
		Msg:      msgBytes,
		Funds:    sdk.NewCoins(),
	}

	return &msg
}

func CreateWithdrawAndSwapMsg(sender, contract, amount, minAmountOut, denom string, poolID uint64) sdk.Msg {
	// Populate the struct with your variables
	msgStruct := ExecuteMsg{
		VaultExtension: &ExtensionExecuteMsg{
			Vaultenator: &VaultenatorExtensionExecuteMsg{
				WithdrawAndSwap: &WithdrawAndSwap{
					WithdrawAmount: amount,
					MinAmountOut:   minAmountOut,
					PoolID:         poolID,
					Denom:          denom,
				},
			},
		},
	}

	// Marshal the struct into JSON
	msgBytes, err := json.Marshal(msgStruct)
	if err != nil {
		panic(err)
	}
	// Generate the swap message
	msg := wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contract,
		Msg:      msgBytes,
		Funds:    sdk.NewCoins(),
	}

	return &msg
}

func CreateRepayMsg(sender, contract string, token sdk.Coin) sdk.Msg {
	// Populate the struct with your variables
	msgStruct := ExecuteMsg{
		VaultExtension: &ExtensionExecuteMsg{
			Vaultenator: &VaultenatorExtensionExecuteMsg{
				Repay: &Repay{},
			},
		},
	}

	// Marshal the struct into JSON
	msgBytes, err := json.Marshal(msgStruct)
	if err != nil {
		panic(err)
	}
	// Generate the swap message
	msg := wasmtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contract,
		Msg:      msgBytes,
		Funds:    sdk.NewCoins(token),
	}

	return &msg
}
