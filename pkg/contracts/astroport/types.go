package astroport

import (
	"encoding/json"
	"fmt"
)

// SwapMessage represents the payload structure for the contract call.
type SwapMessage struct {
	Swap SwapDetails `json:"swap"`
}

// SwapDetails holds the details for the swap action.
type SwapDetails struct {
	OfferAsset   Asset      `json:"offer_asset"`
	AskAssetInfo *AssetInfo `json:"ask_asset_info,omitempty"`
	BeliefPrice  string     `json:"belief_price,omitempty"`
	MaxSpread    string     `json:"max_spread,omitempty"`
	To           string     `json:"to,omitempty"`
}

// Asset contains the offered asset information.
type Asset struct {
	Info   AssetInfo `json:"info"`
	Amount string    `json:"amount"`
}

// AssetInfo represents the information of the token being offered.
type AssetInfo struct {
	Token       *Token       `json:"token,omitempty"`
	NativeToken *NativeToken `json:"native_token,omitempty"`
}

// Token represents the contract address of a token.
type Token struct {
	ContractAddr string `json:"contract_addr"`
}

// NativeToken holds the denomination of the native token.
type NativeToken struct {
	Denom string `json:"denom"`
}

// MarshalJSON customizes JSON serialization for AssetInfo to handle enum-like behavior.
// MarshalJSON customizes JSON serialization for AssetInfo to handle both Token and NativeToken types.
func (a AssetInfo) MarshalJSON() ([]byte, error) {
	if a.Token != nil && a.NativeToken != nil {
		return nil, fmt.Errorf("both Token and NativeToken cannot be set simultaneously in AssetInfo")
	}

	if a.Token != nil {
		return json.Marshal(struct {
			Token *Token `json:"token"`
		}{
			Token: a.Token,
		})
	}
	if a.NativeToken != nil {
		return json.Marshal(struct {
			NativeToken *NativeToken `json:"native_token"`
		}{
			NativeToken: a.NativeToken,
		})
	}

	// Return an error if both fields are nil
	return nil, fmt.Errorf("invalid AssetInfo: both Token and NativeToken are nil")
}

// UnmarshalJSON customizes JSON deserialization for AssetInfo to handle both Token and NativeToken types.
func (a *AssetInfo) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as NativeToken
	var nativeToken struct {
		NativeToken *NativeToken `json:"native_token"`
	}
	if err := json.Unmarshal(data, &nativeToken); err == nil && nativeToken.NativeToken != nil {
		a.NativeToken = nativeToken.NativeToken
		a.Token = nil // Clear Token field
		return nil
	}

	// Try to unmarshal as Token
	var token struct {
		Token *Token `json:"token"`
	}
	if err := json.Unmarshal(data, &token); err == nil && token.Token != nil {
		a.Token = token.Token
		a.NativeToken = nil // Clear NativeToken field
		return nil
	}

	return fmt.Errorf("invalid JSON data for AssetInfo: must contain either token or native_token")
}

type SimulationResponse struct {
	ReturnAmount     string `json:"return_amount"`
	SpreadAmount     string `json:"spread_amount"`
	CommissionAmount string `json:"commission_amount"`
}

type PoolResponse struct {
	Assets     []Asset `json:"assets"`
	TotalShare string  `json:"total_share"`
}
