package cw20

// BalanceResponse represents the response structure for a CW20 balance query
type BalanceResponse struct {
	Balance string `json:"balance"`
}

// TokenInfoResponse represents CW20 token metadata
type TokenInfoResponse struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Decimals    int    `json:"decimals"`
	TotalSupply string `json:"total_supply"`
}

// AllowanceResponse represents the allowance for a spender
type AllowanceResponse struct {
	Allowance string `json:"allowance"`
}
