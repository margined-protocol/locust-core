package ibc

import (
	"bytes"
	"encoding/json"
	"fmt"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	neutronfeetypes "github.com/margined-protocol/locust-core/pkg/proto/neutron/feerefunder/types"
	neutrontransfertypes "github.com/margined-protocol/locust-core/pkg/proto/neutron/transfer/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBCConnection manages IBC connections between chains
type IBCConnection struct {
	Transfer      *IBCTransfer
	SourcePrefix  string
	DestPrefix    string
	ForwardPrefix string
}

// IBCTransfer represents an IBC connection between two chains
type IBCTransfer struct {
	SourceChainID string   // Chain ID of the source chain (e.g., "osmosis-1")
	DestChainID   string   // Chain ID of the destination chain (e.g., "neutron-1")
	Channel       string   // Channel ID on the source chain
	Port          string   // Port ID on the source chain (usually "transfer")
	Forward       *Forward // Forwarding address on the destination chain (optional)
}

// Forward represents a forward  between two chains
type Forward struct {
	ChainID  string // Chain ID of the middle chain
	Receiver string // TODO: remove this and just use the receiver address
	Port     string
	Channel  string
}

// Structs for IBC memo formatting - defined flat to avoid nesting
type WasmExecInfo struct {
	Contract string          `json:"contract"`
	Msg      json.RawMessage `json:"msg"`
}

type WasmExecMemo struct {
	Wasm WasmExecInfo `json:"wasm"`
}

type ForwardNextInfo struct {
	Wasm WasmExecInfo `json:"wasm"`
}

type ForwardInfo struct {
	Channel  string           `json:"channel"`
	Next     *ForwardNextInfo `json:"next,omitempty"`
	Port     string           `json:"port"`
	Receiver string           `json:"receiver"`
	Retries  *int             `json:"retries,omitempty"`
	Timeout  *int64           `json:"timeout,omitempty"`
}

type ForwardMemo struct {
	Forward ForwardInfo `json:"forward"`
}

// orderedMarshalJSON is a helper function to marshal a struct with explicit field ordering
func orderedMarshalJSON(orderedFields []struct {
	Key   string
	Value interface{}
},
) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.WriteString("{"); err != nil {
		return nil, err
	}

	for i, field := range orderedFields {
		if i > 0 {
			if _, err := buf.WriteString(","); err != nil {
				return nil, err
			}
		}

		// Marshal the field key
		keyJSON, err := json.Marshal(field.Key)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(keyJSON); err != nil {
			return nil, err
		}

		if _, err := buf.WriteString(":"); err != nil {
			return nil, err
		}

		// Marshal the field value
		valueJSON, err := json.Marshal(field.Value)
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(valueJSON); err != nil {
			return nil, err
		}
	}

	if _, err := buf.WriteString("}"); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// MarshalJSON for WasmExecInfo
func (w WasmExecInfo) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON([]struct {
		Key   string
		Value interface{}
	}{
		{"contract", w.Contract},
		{"msg", w.Msg},
	})
}

// MarshalJSON for ForwardNextInfo
func (f ForwardNextInfo) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON([]struct {
		Key   string
		Value interface{}
	}{
		{"wasm", f.Wasm},
	})
}

// MarshalJSON for ForwardMemo
func (f ForwardMemo) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON([]struct {
		Key   string
		Value interface{}
	}{
		{"forward", f.Forward},
	})
}

// MarshalJSON for WasmExecMemo
func (w WasmExecMemo) MarshalJSON() ([]byte, error) {
	return orderedMarshalJSON([]struct {
		Key   string
		Value interface{}
	}{
		{"wasm", w.Wasm},
	})
}

func CreateTransferMsg(port, channel, memo, sender, receiver string, token sdk.Coin, height uint64) sdk.Msg {
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		token,
		sender, receiver,
		clienttypes.NewHeight(1, height+10),
		0,
		memo,
	)
}

// CreateTransferWithMemo creates an IBC transfer message with a memo if forwarding
func CreateTransferWithMemo(
	conn *IBCTransfer,
	sourceChainID, destChainID string,
	coin sdk.Coin,
	blockHeight uint64,
	sender, receiver string,
) (sdk.Msg, error) {
	// Create memo and determine receiver based on connection type
	memo, receiver, err := CreateForwardMemo(conn, receiver, destChainID)
	if err != nil {
		return nil, err
	}

	var msg sdk.Msg
	if sourceChainID == "neutron-1" {
		msg = &neutrontransfertypes.MsgTransfer{
			SourcePort:    conn.Port,
			SourceChannel: conn.Channel,
			Token:         coin,
			Sender:        sender,
			Receiver:      receiver,
			TimeoutHeight: clienttypes.NewHeight(1, blockHeight+10),
			Memo:          memo,
			Fee: neutronfeetypes.Fee{
				AckFee: sdk.Coins{
					sdk.NewCoin("untrn", sdkmath.NewInt(100000)),
				},
				TimeoutFee: sdk.Coins{
					sdk.NewCoin("untrn", sdkmath.NewInt(100000)),
				},
			},
		}
	} else {
		// Create the IBC transfer message with the memo
		msg = CreateTransferMsg(
			conn.Port,
			conn.Channel,
			memo,
			sender,
			receiver,
			coin,
			blockHeight,
		)
	}

	return msg, nil
}

// createForwardMemo creates a properly formatted memo for IBC transfers based on connection type
func CreateForwardMemo(
	conn *IBCTransfer,
	receiver string,
	destChainID string,
) (string, string, error) {
	// Validate inputs
	if conn == nil {
		return "", "", fmt.Errorf("connection configuration is nil")
	}

	var memo, finalReceiver string

	// For forwarded IBC transfers (passing through an intermediary chain)
	if conn.Forward != nil {
		finalReceiver = conn.Forward.Receiver

		// Validate forward receiver is not empty
		if finalReceiver == "" {
			return "", "", fmt.Errorf("forward receiver address is empty")
		}

		forwardInfo := ForwardInfo{
			Channel:  conn.Forward.Channel,
			Port:     "transfer",
			Receiver: receiver,
		}

		forwardMemo := ForwardMemo{
			Forward: forwardInfo,
		}

		// Marshal the flat struct to JSON
		// memoBytes, err := json.Marshal(forwardMemo)
		memoBytes, err := forwardMemo.MarshalJSON()
		if err != nil {
			return "", "", fmt.Errorf("failed to marshal forward memo: %w", err)
		}
		memo = string(memoBytes)
	} else {
		finalReceiver = receiver

		memo = ""
	}

	return memo, finalReceiver, nil
}
