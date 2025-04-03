package astroport

import (
	"encoding/json"
	"fmt"

	wasmdtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdkmath "cosmossdk.io/math"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

// CreatePCLSwapMessage constructs a PCL Swap message for the Astroport module, supporting both Token and NativeToken types.
func CreatePCLSwapMessage(amount, denom, beliefPrice, maxSpread, to, contractAddr string, isToken bool) (*SwapMessage, error) {
	// Validate mandatory fields
	if amount == "" {
		return nil, fmt.Errorf("amount must be a non-empty string")
	}

	// Determine which type of AssetInfo to create (Token or NativeToken)
	var assetInfo AssetInfo
	if isToken {
		if contractAddr == "" {
			return nil, fmt.Errorf("contract address must be provided for token type")
		}
		// Create AssetInfo for a Token
		assetInfo = AssetInfo{
			Token: &Token{
				ContractAddr: contractAddr,
			},
		}
	} else {
		if denom == "" {
			return nil, fmt.Errorf("denom must be provided for native token type")
		}
		// Create AssetInfo for a NativeToken
		assetInfo = AssetInfo{
			NativeToken: &NativeToken{
				Denom: denom,
			},
		}
	}

	// Ensure that at least one of the AssetInfo fields is set
	if assetInfo.Token == nil && assetInfo.NativeToken == nil {
		return nil, fmt.Errorf("invalid AssetInfo: both Token and NativeToken are nil")
	}

	// Create the Asset object
	offerAsset := Asset{
		Info:   assetInfo,
		Amount: amount,
	}

	// Create the SwapMessage using the provided parameters
	msg := &SwapMessage{
		Swap: SwapDetails{
			OfferAsset:  offerAsset,
			BeliefPrice: beliefPrice,
			MaxSpread:   maxSpread,
			To:          to,
		},
	}

	return msg, nil
}

// CreateAstroportSwapMsg constructs a MsgExecuteContract for an Astroport swap with the given parameters.
func CreateAstroportSwapMsg(sender, contractAddress, offerAsset, beliefPrice, maxSpread string, amountToBuy sdkmath.Int) (*wasmdtypes.MsgExecuteContract, error) {
	// Generate the Astroport swap message
	astroportSwapMsg, err := CreatePCLSwapMessage(
		amountToBuy.String(), // Amount to buy as a string
		offerAsset,           // Offer asset (Denom)
		beliefPrice,          // Belief price
		maxSpread,            // Max spread (can be empty)
		"",                   // To address (can be empty)
		"",                   // Contract address (not used for NativeToken)
		false,                // isToken (false for NativeToken)
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Astroport swap message: %w", err)
	}

	// Convert the Astroport swap message to JSON bytes
	buyMsgBytes, err := json.Marshal(astroportSwapMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Astroport swap message: %w", err)
	}

	// Construct the MsgExecuteContract with the generated message and funds
	msgExecuteContract := &wasmdtypes.MsgExecuteContract{
		Sender:   sender,
		Contract: contractAddress,
		Msg:      buyMsgBytes,
		Funds: []sdktypes.Coin{
			{
				Denom:  offerAsset,
				Amount: amountToBuy,
			},
		},
	}

	return msgExecuteContract, nil
}
