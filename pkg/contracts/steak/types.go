package steak

import (
	"github.com/cosmos/cosmos-sdk/types"
)

// osmo1s3l0lcqc7tu0vpj6wdjz9wqpxv8nk6eraevje4fuwkyjnwuy82qsx3lduv - backbone
type BackboneStateResponse struct {
	TotalUsteak   string       `json:"total_usteak"`
	TotalNative   string       `json:"total_native"`
	ExchangeRate  string       `json:"exchange_rate"`
	UnlockedCoins []types.Coin `json:"unlocked_coins"`
}

// osmo1dv8wz09tckslr2wy5z86r46dxvegylhpt97r9yd6qc3kyc6tv42qa89dr9 - eris
type ErisStateResponse struct {
	TotalUsteak  string       `json:"total_usteak"`
	TotalNative  string       `json:"total_native"`
	ExchangeRate string       `json:"exchange_rate"`
	UnlockCoins  []types.Coin `json:"unlocked_coin"`
	Unbonding    string       `json:"unbonding"`
	Available    string       `json:"available"`
	TVLUtoken    string       `json:"tvl_utoken"`
}
